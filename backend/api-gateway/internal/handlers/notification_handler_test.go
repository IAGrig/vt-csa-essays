package handlers_test

import (
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

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

func TestNotificationHandler_GetUserNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         interface{}
		setupMock      func(*mocks.MockNotificationClient)
		expectedStatus int
		expectedLength int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful get user notifications",
			userID: int64(123),
			setupMock: func(mockClient *mocks.MockNotificationClient) {
				mockClient.On("GetByUserID", mock.Anything, &pb.GetByUserIDRequest{
					UserId: 123,
				}).Return([]*pb.NotificationResponse{
					{
						NotificationId: 1,
						UserId:         123,
						Content:        "Your essay has been reviewed!",
						IsRead:         false,
						CreatedAt:      1234567890,
					},
					{
						NotificationId: 2,
						UserId:         123,
						Content:        "New comment on your essay",
						IsRead:         true,
						CreatedAt:      1234567891,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLength: 2,
		},
		{
			name:   "no notifications for user",
			userID: int64(456),
			setupMock: func(mockClient *mocks.MockNotificationClient) {
				mockClient.On("GetByUserID", mock.Anything, &pb.GetByUserIDRequest{
					UserId: 456,
				}).Return([]*pb.NotificationResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLength: 0,
		},
		{
			name:           "missing authentication",
			userID:         nil,
			setupMock:      func(mockClient *mocks.MockNotificationClient) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "authentication required",
			},
		},
		{
			name:   "notification service error",
			userID: int64(123),
			setupMock: func(mockClient *mocks.MockNotificationClient) {
				mockClient.On("GetByUserID", mock.Anything, mock.Anything).
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
			mockNotificationClient := new(mocks.MockNotificationClient)
			tt.setupMock(mockNotificationClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewNotificationHandler(mockNotificationClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodGet, "/notifications", nil)
			require.NoError(t, err)

			c.Request = req

			if tt.userID != nil {
				c.Set("user_id", tt.userID)
			}

			handler.GetUserNotifications(c)

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

			mockNotificationClient.AssertExpectations(t)
		})
	}
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		notificationId string
		setupMock      func(*mocks.MockNotificationClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "successful mark as read",
			notificationId: "1",
			setupMock: func(mockClient *mocks.MockNotificationClient) {
				mockClient.On("MarkAsRead", mock.Anything, &pb.MarkAsReadRequest{
					NotificationId: 1,
				}).Return(&pb.MarkAsReadResponse{
					Success: true,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name:           "invalid notification ID",
			notificationId: "invalid",
			setupMock:      func(mockClient *mocks.MockNotificationClient) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid notification ID",
			},
		},
		{
			name:           "notification not found",
			notificationId: "999",
			setupMock: func(mockClient *mocks.MockNotificationClient) {
				mockClient.On("MarkAsRead", mock.Anything, &pb.MarkAsReadRequest{
					NotificationId: 999,
				}).Return(&pb.MarkAsReadResponse{
					Success: false,
				}, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "notification not found",
			},
		},
		{
			name:           "service error",
			notificationId: "1",
			setupMock: func(mockClient *mocks.MockNotificationClient) {
				mockClient.On("MarkAsRead", mock.Anything, mock.Anything).
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
			mockNotificationClient := new(mocks.MockNotificationClient)
			tt.setupMock(mockNotificationClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewNotificationHandler(mockNotificationClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodPost, "/notifications/"+tt.notificationId+"/read", nil)
			require.NoError(t, err)

			c.Request = req
			c.Params = gin.Params{gin.Param{Key: "notificationId", Value: tt.notificationId}}

			handler.MarkAsRead(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockNotificationClient.AssertExpectations(t)
		})
	}
}

func TestNotificationHandler_MarkAllAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         interface{}
		setupMock      func(*mocks.MockNotificationClient)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful mark all as read",
			userID: int64(123),
			setupMock: func(mockClient *mocks.MockNotificationClient) {
				mockClient.On("MarkAllAsRead", mock.Anything, &pb.MarkAllAsReadRequest{
					UserId: 123,
				}).Return(&pb.MarkAllAsReadResponse{
					Success: true,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name:           "missing authentication",
			userID:         nil,
			setupMock:      func(mockClient *mocks.MockNotificationClient) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "authentication required",
			},
		},
		{
			name:   "service returns failure",
			userID: int64(123),
			setupMock: func(mockClient *mocks.MockNotificationClient) {
				mockClient.On("MarkAllAsRead", mock.Anything, &pb.MarkAllAsReadRequest{
					UserId: 123,
				}).Return(&pb.MarkAllAsReadResponse{
					Success: false,
				}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "failed to mark notifications as read",
			},
		},
		{
			name:   "service error",
			userID: int64(123),
			setupMock: func(mockClient *mocks.MockNotificationClient) {
				mockClient.On("MarkAllAsRead", mock.Anything, mock.Anything).
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
			mockNotificationClient := new(mocks.MockNotificationClient)
			tt.setupMock(mockNotificationClient)

			logger := logging.NewEmptyLogger()
			handler := handlers.NewNotificationHandler(mockNotificationClient, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(http.MethodPost, "/notifications/read-all", nil)
			require.NoError(t, err)

			c.Request = req

			if tt.userID != nil {
				c.Set("user_id", tt.userID)
			}

			handler.MarkAllAsRead(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockNotificationClient.AssertExpectations(t)
		})
	}
}
