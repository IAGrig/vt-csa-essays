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

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

func TestReviewHandler_CreateReview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    []byte
		username       interface{}
		setupMock      func(*mocks.MockReviewClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful review creation",
			requestBody: []byte(`{
				"essay_id": 123,
				"rank": 1,
				"content": "Great essay!"
			}`),
			username: "reviewer1",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("CreateReview", mock.Anything, &pb.ReviewAddRequest{
					EssayId: 123,
					Rank:    1,
					Content: "Great essay!",
					Author:  "reviewer1",
				}).Return(&pb.ReviewResponse{
					Id:      1,
					EssayId: 123,
					Rank:    1,
					Content: "Great essay!",
					Author:  "reviewer1",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":       float64(1),
				"essay_id": float64(123),
				"rank":     float64(1),
				"content":  "Great essay!",
				"author":   "reviewer1",
			},
		},
		{
			name: "missing authentication",
			requestBody: []byte(`{
				"essay_id": 123,
				"rank": 1,
				"content": "Great essay!"
			}`),
			username: nil,
			setupMock: func(mockClient *mocks.MockReviewClient) {
				// no call expected
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "authentication required",
			},
		},
		{
			name:        "invalid json",
			requestBody: []byte("invalid json"),
			username:    "reviewer1",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				// no call expected
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing required fields",
			requestBody: []byte(`{
				"rank": 1,
				"content": "Great essay!"
			}`),
			username: "reviewer1",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				// no call expected
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "review service error",
			requestBody: []byte(`{
				"essay_id": 123,
				"rank": 1,
				"content": "Great essay!"
			}`),
			username: "reviewer1",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("CreateReview", mock.Anything, mock.Anything).
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
			mockReviewClient := new(mocks.MockReviewClient)
			tt.setupMock(mockReviewClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewReviewHandler(mockReviewClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodPost, "/reviews", bytes.NewBuffer(tt.requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			c.Request = req

			if tt.username != nil {
				c.Set("username", tt.username)
			}

			handler.CreateReview(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockReviewClient.AssertExpectations(t)
		})
	}
}

func TestReviewHandler_GetAllReviews(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(*mocks.MockReviewClient)
		expectedStatus int
		expectedLength int
	}{
		{
			name: "successful get all reviews",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("GetAllReviews", mock.Anything, &pb.EmptyRequest{}).
					Return([]*pb.ReviewResponse{
						{
							Id:      1,
							EssayId: 123,
							Rank:    1,
							Content: "Great essay!",
							Author:  "reviewer1",
						},
						{
							Id:      2,
							EssayId: 123,
							Rank:    2,
							Content: "Good essay",
							Author:  "reviewer2",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLength: 2,
		},
		{
			name: "empty reviews list",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("GetAllReviews", mock.Anything, mock.Anything).
					Return([]*pb.ReviewResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLength: 0,
		},
		{
			name: "service error",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("GetAllReviews", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReviewClient := new(mocks.MockReviewClient)
			tt.setupMock(mockReviewClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewReviewHandler(mockReviewClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodGet, "/reviews", nil)
			require.NoError(t, err)

			c.Request = req

			handler.GetAllReviews(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response, tt.expectedLength)
			}

			mockReviewClient.AssertExpectations(t)
		})
	}
}

func TestReviewHandler_GetByEssayId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		essayId        string
		setupMock      func(*mocks.MockReviewClient)
		expectedStatus int
		expectedLength int
		expectedBody   map[string]interface{}
	}{
		{
			name:    "reviews found for essay",
			essayId: "123",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("GetByEssayId", mock.Anything, &pb.GetByEssayIdRequest{
					EssayId: 123,
				}).Return([]*pb.ReviewResponse{
					{
						Id:      1,
						EssayId: 123,
						Rank:    1,
						Content: "Great essay!",
						Author:  "reviewer1",
					},
					{
						Id:      2,
						EssayId: 123,
						Rank:    2,
						Content: "Good essay",
						Author:  "reviewer2",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLength: 2,
		},
		{
			name:    "no reviews for essay",
			essayId: "456",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("GetByEssayId", mock.Anything, &pb.GetByEssayIdRequest{
					EssayId: 456,
				}).Return([]*pb.ReviewResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLength: 0,
		},
		{
			name:    "invalid essay ID",
			essayId: "invalid",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				// no call expected
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid essay ID",
			},
		},
		{
			name:    "reviews not found",
			essayId: "999",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("GetByEssayId", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "reviews not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReviewClient := new(mocks.MockReviewClient)
			tt.setupMock(mockReviewClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewReviewHandler(mockReviewClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodGet, "/reviews/"+tt.essayId, nil)
			require.NoError(t, err)

			c.Request = req
			c.Params = gin.Params{gin.Param{Key: "essayId", Value: tt.essayId}}

			handler.GetByEssayId(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response, tt.expectedLength)
			} else if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody["error"], response["error"])
			}

			mockReviewClient.AssertExpectations(t)
		})
	}
}

func TestReviewHandler_RemoveById(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		reviewId       string
		setupMock      func(*mocks.MockReviewClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:     "successful deletion",
			reviewId: "1",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("RemoveById", mock.Anything, &pb.RemoveByIdRequest{
					Id: 1,
				}).Return(&pb.ReviewResponse{
					Id:      1,
					EssayId: 123,
					Rank:    1,
					Content: "Deleted review",
					Author:  "reviewer1",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":       float64(1),
				"essay_id": float64(123),
				"rank":     float64(1),
				"content":  "Deleted review",
				"author":   "reviewer1",
			},
		},
		{
			name:     "invalid review ID",
			reviewId: "invalid",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				// no call expected
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid review ID",
			},
		},
		{
			name:     "review not found",
			reviewId: "999",
			setupMock: func(mockClient *mocks.MockReviewClient) {
				mockClient.On("RemoveById", mock.Anything, &pb.RemoveByIdRequest{
					Id: 999,
				}).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReviewClient := new(mocks.MockReviewClient)
			tt.setupMock(mockReviewClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewReviewHandler(mockReviewClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodDelete, "/reviews/"+tt.reviewId, nil)
			require.NoError(t, err)

			c.Request = req
			c.Params = gin.Params{gin.Param{Key: "reviewId", Value: tt.reviewId}}

			handler.RemoveById(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockReviewClient.AssertExpectations(t)
		})
	}
}
