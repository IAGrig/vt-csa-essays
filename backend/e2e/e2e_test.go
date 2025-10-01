package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type httpResp struct {
	body   []byte
	status int
}

func httpGet(t *testing.T, url string, headers map[string]string) (httpResp, error) {
	t.Helper()

	t.Logf("GET %s", url)
	t.Logf("Headers: %s", headers)
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return httpResp{}, fmt.Errorf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("Response %d: %s", resp.StatusCode, string(body))
	return httpResp{body, resp.StatusCode}, nil
}

func httpPost(t *testing.T, url string, data string, headers map[string]string) (httpResp, error) {
	t.Helper()

	t.Logf("POST %s %s", url, data)
	t.Logf("Headers: %s", headers)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return httpResp{}, fmt.Errorf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("Response %d: %s", resp.StatusCode, string(body))
	return httpResp{body, resp.StatusCode}, nil
}

func waitForHealthy(t *testing.T, url string, timeout time.Duration) error {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			t.Log("API is healthy")
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		t.Log("Waiting for API ...")
		time.Sleep(time.Second)
	}
	return fmt.Errorf("API not healthy after %v", timeout)
}

func dockerUp(t *testing.T) error {
	t.Helper()

	cmd := exec.Command("docker", "compose", "-f", "../../docker-compose.yaml", "up", "--build", "-d")
	out, err := cmd.CombinedOutput()
	t.Logf("docker compose up output:\n%s", out)
	if err != nil {
		return fmt.Errorf("docker compose up failed: %v", err)
	}
	return nil
}

func dockerDown(t *testing.T) error {
	t.Helper()

	cmd := exec.Command("docker", "compose", "-f", "../../docker-compose.yaml", "down", "-v", "--remove-orphans")
	out, err := cmd.CombinedOutput()
	t.Logf("docker compose down output:\n%s", out)
	if err != nil {
		return fmt.Errorf("docker compose down failed: %v", err)
	}
	return nil
}

var errors []string

func checkError(t *testing.T, operation string, err error) {
	t.Helper()

	if err != nil {
		errors = append(errors, fmt.Sprintf("%s: %v", operation, err))
		t.Errorf("%s failed: %v", operation, err)
	}
}

func checkStatus(t *testing.T, operation string, resp httpResp, expectedStatus int) {
	t.Helper()

	if resp.status != expectedStatus {
		msg := fmt.Sprintf("%s: expected status %d, got %d", operation, expectedStatus, resp.status)
		errors = append(errors, msg)
		t.Error(msg)
	}
}

func checkNotEmpty(t *testing.T, operation string, data interface{}) {
	t.Helper()

	switch v := data.(type) {
	case []interface{}:
		if len(v) == 0 {
			msg := fmt.Sprintf("%s: expected non-empty array, got empty", operation)
			errors = append(errors, msg)
			t.Error(msg)
		}
	case map[string]interface{}:
		if len(v) == 0 {
			msg := fmt.Sprintf("%s: expected non-empty object, got empty", operation)
			errors = append(errors, msg)
			t.Error(msg)
		}
	case nil:
		msg := fmt.Sprintf("%s: expected data, got null", operation)
		errors = append(errors, msg)
		t.Error(msg)
	}
}

func TestE2EHappyPath(t *testing.T) {
	if err := dockerUp(t); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer func() {
		if err := dockerDown(t); err != nil {
			t.Logf("Teardown warning: %v", err)
		}
	}()

	apiURL := "http://localhost:8080"
	if err := waitForHealthy(t, apiURL+"/api/essays", 60*time.Second); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	// register users
	resp, err := httpPost(t, apiURL+"/api/auth/register", `{"username":"e2e_user1","password":"pw"}`, nil)
	checkError(t, "Register user1", err)
	checkStatus(t, "Register user1", resp, http.StatusCreated)

	var user1Id int
	if err == nil && resp.status == http.StatusCreated {
		var register1Resp map[string]interface{}
		err := json.Unmarshal(resp.body, &register1Resp)
		if err != nil {
			checkError(t, "Parse register1 response", err)
		}
		idFloat, ok := register1Resp["id"].(float64)
		if !ok {
			t.Error("Error: non-numeric type of user ID")
		} else {
			user1Id = int(idFloat)
		}
	}

	resp, err = httpPost(t, apiURL+"/api/auth/register", `{"username":"e2e_user2","password":"pw"}`, nil)
	checkError(t, "Register user2", err)
	checkStatus(t, "Register user2", resp, http.StatusCreated)

	// login user1
	resp, err = httpPost(t, apiURL+"/api/auth/login", `{"username":"e2e_user1","password":"pw"}`, nil)
	checkError(t, "Login user1", err)
	checkStatus(t, "Login user1", resp, http.StatusOK)

	var login1 map[string]interface{}
	if err == nil && resp.status == http.StatusOK {
		if err := json.Unmarshal(resp.body, &login1); err != nil {
			checkError(t, "Parse login1 response", err)
		} else {
			checkNotEmpty(t, "Login response", login1)
		}
	}

	token1, _ := login1["access_token"].(string)
	headers1 := map[string]string{"Authorization": "Bearer " + token1}

	// user1 creates essay
	var essay map[string]interface{}
	var essayID string
	if token1 != "" {
		resp, err = httpPost(t, apiURL+"/api/essays", `{"content":"Test essay text"}`, headers1)
		checkError(t, "Create essay", err)
		checkStatus(t, "Create essay", resp, http.StatusCreated)

		if err == nil && resp.status == http.StatusCreated {
			if err := json.Unmarshal(resp.body, &essay); err != nil {
				checkError(t, "Parse essay response", err)
			} else {
				checkNotEmpty(t, "Essay response", essay)
				essayID = fmt.Sprint(essay["id"])
			}
		}
	}

	// list essays, should not be null
	resp, err = httpGet(t, apiURL+"/api/essays", nil)
	checkError(t, "List essays", err)
	checkStatus(t, "List essays", resp, http.StatusOK)

	if err == nil && resp.status == http.StatusOK {
		var essays []interface{}
		if err := json.Unmarshal(resp.body, &essays); err != nil {
			checkError(t, "Parse essays list", err)
		}
		if len(essays) == 0 {
			checkError(t, "Parse essays list", fmt.Errorf("Essays list is empty"))
		}
	}

	// login user2
	resp, err = httpPost(t, apiURL+"/api/auth/login", `{"username":"e2e_user2","password":"pw"}`, nil)
	checkError(t, "Login user2", err)
	checkStatus(t, "Login user2", resp, http.StatusOK)

	var login2 map[string]interface{}
	if err == nil && resp.status == http.StatusOK {
		if err := json.Unmarshal(resp.body, &login2); err != nil {
			checkError(t, "Parse login2 response", err)
		} else {
			checkNotEmpty(t, "Login2 response", login2)
		}
	}

	token2, _ := login2["access_token"].(string)
	headers2 := map[string]string{"Authorization": "Bearer " + token2}

	// user2 posts review
	if token2 != "" && essayID != "" {
		resp, err = httpPost(t, apiURL+"/api/reviews",
			fmt.Sprintf(`{"essay_id":%s,"essay_author_id":%d,"rank":1,"content":"Nice!"}`, essayID, user1Id), headers2)
		checkError(t, "Create review", err)
		checkStatus(t, "Create review", resp, http.StatusCreated)
	}

	// check reviews - should not be null
	if essayID != "" {
		resp, err = httpGet(t, apiURL+"/api/reviews/"+essayID, nil)
		checkError(t, "Get reviews", err)
		checkStatus(t, "Get reviews", resp, http.StatusOK)

		if err == nil && resp.status == http.StatusOK {
			var reviews []interface{}
			if err := json.Unmarshal(resp.body, &reviews); err != nil {
				checkError(t, "Parse reviews list", err)
			}
			if len(reviews) == 0 {
				checkError(t, "Parse reviews list", fmt.Errorf("Reviews list is empty"))
			}
		}
	}

	// wait for notifications (kafka)
	t.Log("waiting 10s for notifications to propagate")
	time.Sleep(10 * time.Second)

	if token1 != "" {
		resp, err = httpGet(t, apiURL+"/api/notifications", headers1)
		checkError(t, "Get notifications", err)
		checkStatus(t, "Get notifications", resp, http.StatusOK)

		if err == nil && resp.status == http.StatusOK {
			var notifications []interface{}
			if err := json.Unmarshal(resp.body, &notifications); err != nil {
				checkError(t, "Parse notifications list", err)
			}
			if len(notifications) == 0 {
				checkError(t, "Parse notifications list", fmt.Errorf("Notifications list is empty"))
			}
		}
	}

	if len(errors) > 0 {
		t.Errorf("Test completed with %d errors:\n%s", len(errors), strings.Join(errors, "\n"))
	} else {
		t.Log("All test steps completed successfully!")
	}
}
