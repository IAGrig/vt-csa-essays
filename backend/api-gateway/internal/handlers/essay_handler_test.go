package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients/mocks"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/handlers"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
	reviewPb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

func TestEssayHandler_CreateEssay(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    []byte
		username       interface{}
		setupMock      func(*mocks.MockEssayClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful essay creation",
			requestBody: []byte(`{
				"content": "This is a test essay content"
			}`),
			username: "testuser",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("CreateEssay", mock.Anything, &pb.EssayAddRequest{
					Content: "This is a test essay content",
					Author:  "testuser",
				}).Return(&pb.EssayResponse{
					Id:       1,
					Content:  "This is a test essay content",
					Author:   "testuser",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":      float64(1),
				"content": "This is a test essay content",
				"author":  "testuser",
			},
		},
		{
			name: "missing authentication",
			requestBody: []byte(`{
				"content": "This is a test essay content"
			}`),
			username: nil,
			setupMock: func(mockClient *mocks.MockEssayClient) {
				// no call expected
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "authentication required",
			},
		},
		{
			name: "invalid json",
			requestBody: []byte("invalid json"),
			username: "testuser",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				// no call expected
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing content",
			requestBody: []byte(`{}`),
			username: "testuser",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				// no call expected
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "essay service error",
			requestBody: []byte(`{
				"content": "This is a test essay content"
			}`),
			username: "testuser",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("CreateEssay", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": assert.AnError.Error(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEssayClient := new(mocks.MockEssayClient)
			tt.setupMock(mockEssayClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewEssayHandler(mockEssayClient, logger)

			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodPost, "/essays", bytes.NewBuffer(tt.requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			c.Request = req

			if tt.username != nil {
				c.Set("username", tt.username)
			}

			handler.CreateEssay(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockEssayClient.AssertExpectations(t)
		})
	}

}

func TestEssayHandler_GetEssay(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authorname     string
		setupMock      func(*mocks.MockEssayClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "essay found",
			authorname: "testauthor",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("GetEssay", mock.Anything, &pb.GetByAuthorNameRequest{
					Authorname: "testauthor",
				}).Return(&pb.EssayWithReviewsResponse{
					Id:      1,
					Content: "Test content",
					Author:  "testauthor",
					Reviews: []*reviewPb.ReviewResponse{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":      float64(1),
				"content": "Test content",
				"author":  "testauthor",
				"reviews": []interface{}{},
			},
		},
		{
			name:       "essay not found",
			authorname: "nonexistent",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("GetEssay", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "essay not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEssayClient := new(mocks.MockEssayClient)
			tt.setupMock(mockEssayClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewEssayHandler(mockEssayClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodGet, "/essays/"+tt.authorname, nil)
			require.NoError(t, err)

			c.Request = req
			c.Params = gin.Params{gin.Param{Key: "authorname", Value: tt.authorname}}

			handler.GetEssay(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockEssayClient.AssertExpectations(t)
		})
	}
}

func TestEssayHandler_GetAllEssays(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*mocks.MockEssayClient)
		expectedStatus int
		expectedLength int
	}{
		{
			name:        "get all essays",
			queryParams: "",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("GetAllEssays", mock.Anything, &pb.EmptyRequest{}).
					Return([]*pb.EssayResponse{
						{Id: 1, Content: "Content 1", Author: "user1"},
						{Id: 2, Content: "Content 2", Author: "user2"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLength: 2,
		},
		{
			name:        "search essays",
			queryParams: "?search=test",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("SearchEssays", mock.Anything, &pb.SearchByContentRequest{
					Content: "test",
				}).Return([]*pb.EssayResponse{
					{Id: 1, Content: "Test content", Author: "user1"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLength: 1,
		},
		{
			name:        "empty result",
			queryParams: "?search=nonexistent",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("SearchEssays", mock.Anything, mock.Anything).
					Return([]*pb.EssayResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLength: 0,
		},
		{
			name:        "service error - get all",
			queryParams: "",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("GetAllEssays", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEssayClient := new(mocks.MockEssayClient)
			tt.setupMock(mockEssayClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewEssayHandler(mockEssayClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodGet, "/essays"+tt.queryParams, nil)
			require.NoError(t, err)

			c.Request = req

			handler.GetAllEssays(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response, tt.expectedLength)
			}

			mockEssayClient.AssertExpectations(t)
		})
	}
}

func TestEssayHandler_RemoveEssay(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authorname     string
		username       interface{}
		setupMock      func(*mocks.MockEssayClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "successful deletion - same user",
			authorname: "testuser",
			username:   "testuser",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("DeleteEssay", mock.Anything, &pb.RemoveByAuthorNameRequest{
					Authorname: "testuser",
				}).Return(&pb.EssayResponse{
					Id:      1,
					Content: "Deleted content",
					Author:  "testuser",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":      float64(1),
				"content": "Deleted content",
				"author":  "testuser",
			},
		},
		{
			name:       "forbidden - different user",
			authorname: "otheruser",
			username:   "testuser",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				// no call expected
			},
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]interface{}{
				"error": "you can delete only your own essays",
			},
		},
		{
			name:       "missing authentication",
			authorname: "testuser",
			username:   nil,
			setupMock: func(mockClient *mocks.MockEssayClient) {
				// no call expected
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "authentication required",
			},
		},
		{
			name:       "essay not found",
			authorname: "testuser",
			username:   "testuser",
			setupMock: func(mockClient *mocks.MockEssayClient) {
				mockClient.On("DeleteEssay", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEssayClient := new(mocks.MockEssayClient)
			tt.setupMock(mockEssayClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewEssayHandler(mockEssayClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodDelete, "/essays/"+tt.authorname, nil)
			require.NoError(t, err)

			c.Request = req
			c.Params = gin.Params{gin.Param{Key: "authorname", Value: tt.authorname}}

			if tt.username != nil {
				c.Set("username", tt.username)
			}

			handler.RemoveEssay(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockEssayClient.AssertExpectations(t)
		})
	}
}
