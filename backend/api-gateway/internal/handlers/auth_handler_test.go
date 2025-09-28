package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients/mocks"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mocks.MockAuthClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful registration",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func(mockClient *mocks.MockAuthClient) {
				mockClient.On("Register", mock.Anything, &pb.UserRegisterRequest{
					Username: "testuser",
					Password: "password123",
				}).Return(&pb.UserResponse{
					Id:       123,
					Username: "testuser",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":       float64(123),
				"username": "testuser",
			},
		},
		{
			name: "invalid json",
			requestBody: `invalid json`,
			setupMock: func(mockClient *mocks.MockAuthClient) {
				// no calls expected
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "missing username",
			requestBody: map[string]string{
				"password": "password123",
			},
			setupMock: func(mockClient *mocks.MockAuthClient) {
				// no calls expected
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Key: 'Username' Error:Field validation for 'Username' failed on the 'required' tag",
			},
		},
		{
			name: "auth service error",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func(mockClient *mocks.MockAuthClient) {
				mockClient.On("Register", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": assert.AnError.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthClient := new(mocks.MockAuthClient)
			tt.setupMock(mockAuthClient)

			handler := handlers.NewAuthHandler(mockAuthClient)

			router := gin.New()
			router.POST("/register", handler.Register)

			var reqBody []byte
			switch body := tt.requestBody.(type) {
			case string:
				reqBody = []byte(body)
			default:
				reqBody, _ = json.Marshal(body)
			}

			req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedBody != nil {
				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mocks.MockAuthClient)
		expectedStatus int
		checkCookies   bool
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func(mockClient *mocks.MockAuthClient) {
				mockClient.On("Login", mock.Anything, &pb.UserLoginRequest{
					Username: "testuser",
					Password: "password123",
				}).Return(&pb.AuthTokensResponse{
					AccessToken:  "access_token_123",
					RefreshToken: "refresh_token_123",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkCookies:   true,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setupMock: func(mockClient *mocks.MockAuthClient) {
				mockClient.On("Login", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid credentials",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthClient := new(mocks.MockAuthClient)
			tt.setupMock(mockAuthClient)

			handler := handlers.NewAuthHandler(mockAuthClient)

			router := gin.New()
			router.POST("/login", handler.Login)

			reqBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkCookies {
				cookies := w.Result().Cookies()
				assert.Len(t, cookies, 1)
				assert.Equal(t, "refresh_token", cookies[0].Name)
				assert.Equal(t, "refresh_token_123", cookies[0].Value)
				assert.Equal(t, "/api/auth/refresh", cookies[0].Path)
				assert.True(t, cookies[0].HttpOnly)

				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "access_token_123", response["access_token"])
			}

			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		refreshToken   string
		setupMock      func(*mocks.MockAuthClient)
		expectedStatus int
	}{
		{
			name:         "successful token refresh",
			refreshToken: "valid_refresh_token",
			setupMock: func(mockClient *mocks.MockAuthClient) {
				mockClient.On("RefreshToken", mock.Anything, &pb.RefreshTokenRequest{
					RefreshToken: "valid_refresh_token",
				}).Return(&pb.AuthTokensResponse{
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing refresh token",
			refreshToken:   "",
			setupMock:      func(mockClient *mocks.MockAuthClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid_token",
			setupMock: func(mockClient *mocks.MockAuthClient) {
				mockClient.On("RefreshToken", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthClient := new(mocks.MockAuthClient)
			tt.setupMock(mockAuthClient)

			handler := handlers.NewAuthHandler(mockAuthClient)

			router := gin.New()
			router.POST("/refresh", handler.RefreshToken)

			req, err := http.NewRequest(http.MethodPost, "/refresh", nil)
			require.NoError(t, err)

			if tt.refreshToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tt.refreshToken,
				})
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				cookies := w.Result().Cookies()
				assert.Len(t, cookies, 1)
				assert.Equal(t, "new_refresh_token", cookies[0].Value)

				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "new_access_token", response["access_token"])
			}

			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_GetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		username       string
		setupMock      func(*mocks.MockAuthClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:     "user found",
			username: "testuser",
			setupMock: func(mockClient *mocks.MockAuthClient) {
				mockClient.On("GetUser", mock.Anything, &pb.GetByUsernameRequest{
					Username: "testuser",
				}).Return(&pb.UserResponse{
					Id:       123,
					Username: "testuser",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id": float64(123),
				"username": "testuser",
			},
		},
		{
			name:     "user not found",
			username: "nonexistent",
			setupMock: func(mockClient *mocks.MockAuthClient) {
				mockClient.On("GetUser", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "user not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthClient := new(mocks.MockAuthClient)
			tt.setupMock(mockAuthClient)

			handler := handlers.NewAuthHandler(mockAuthClient)

			router := gin.New()
			router.GET("/user/:username", handler.GetUser)

			req, err := http.NewRequest(http.MethodGet, "/user/"+tt.username, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key], "Mismatch for field '%s'", key)
				}
			}

			mockAuthClient.AssertExpectations(t)
		})
	}
}
