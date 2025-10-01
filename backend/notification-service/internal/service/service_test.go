package service

import (
	"context"
	"testing"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/models"
	repoMocks "github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/repository/mocks"
	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

type MinimalServerStream struct {
	ctx          context.Context
	sentMessages []*pb.NotificationResponse
	sendError    error
}

func (m *MinimalServerStream) Send(msg *pb.NotificationResponse) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.sentMessages = append(m.sentMessages, msg)
	return nil
}

func (m *MinimalServerStream) Context() context.Context {
	return m.ctx
}

func (m *MinimalServerStream) SetHeader(md metadata.MD) error  { return nil }
func (m *MinimalServerStream) SendHeader(md metadata.MD) error { return nil }
func (m *MinimalServerStream) SetTrailer(md metadata.MD)       {}
func (m *MinimalServerStream) SendMsg(interface{}) error       { return nil }
func (m *MinimalServerStream) RecvMsg(interface{}) error       { return nil }

func TestNotificationService_GetByUserID(t *testing.T) {
	tests := []struct {
		name          string
		input         *pb.GetByUserIDRequest
		setupMock     func(*repoMocks.MockNotificationRepository)
		expectedCount int
		sendError     error
		expectedError bool
	}{
		{
			name:  "success - streams user notifications",
			input: &pb.GetByUserIDRequest{UserId: 123},
			setupMock: func(mockRepo *repoMocks.MockNotificationRepository) {
				notifications := []models.Notification{
					{
						NotificationID: 1,
						UserID:         123,
						Content:        "Notification 1",
						IsRead:         false,
					},
					{
						NotificationID: 2,
						UserID:         123,
						Content:        "Notification 2",
						IsRead:         true,
					},
				}
				mockRepo.On("GetByUserID", int64(123)).Return(notifications, nil)
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:  "success - streams empty list when no notifications",
			input: &pb.GetByUserIDRequest{UserId: 456},
			setupMock: func(mockRepo *repoMocks.MockNotificationRepository) {
				mockRepo.On("GetByUserID", int64(456)).Return([]models.Notification{}, nil)
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name:  "error - repository returns error",
			input: &pb.GetByUserIDRequest{UserId: 123},
			setupMock: func(mockRepo *repoMocks.MockNotificationRepository) {
				mockRepo.On("GetByUserID", int64(123)).Return([]models.Notification{}, assert.AnError)
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:  "error - stream send fails",
			input: &pb.GetByUserIDRequest{UserId: 123},
			setupMock: func(mockRepo *repoMocks.MockNotificationRepository) {
				notifications := []models.Notification{
					{
						NotificationID: 1,
						UserID:         123,
						Content:        "Notification 1",
						IsRead:         false,
					},
				}
				mockRepo.On("GetByUserID", int64(123)).Return(notifications, nil)
			},
			sendError:     assert.AnError,
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.MockNotificationRepository)
			tt.setupMock(mockRepo)

			stream := &MinimalServerStream{
				ctx:       context.Background(),
				sendError: tt.sendError,
			}

			logger := logging.NewEmptyLogger()
			service := New(mockRepo, logger)
			err := service.GetByUserID(tt.input, stream)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, stream.sentMessages, tt.expectedCount)

			if tt.expectedCount > 0 {
				assert.Equal(t, int64(1), stream.sentMessages[0].NotificationId)
				assert.Equal(t, "Notification 1", stream.sentMessages[0].Content)
				if tt.expectedCount > 1 {
					assert.Equal(t, int64(2), stream.sentMessages[1].NotificationId)
					assert.Equal(t, "Notification 2", stream.sentMessages[1].Content)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.MarkAsReadRequest
		setupMock      func(*repoMocks.MockNotificationRepository)
		expectedResult *pb.MarkAsReadResponse
		expectedError  bool
	}{
		{
			name:  "success - marks notification as read",
			input: &pb.MarkAsReadRequest{NotificationId: 1},
			setupMock: func(mockRepo *repoMocks.MockNotificationRepository) {
				mockRepo.On("MarkAsRead", int64(1)).Return(nil)
			},
			expectedResult: &pb.MarkAsReadResponse{Success: true},
			expectedError:  false,
		},
		{
			name:  "error - repository returns error",
			input: &pb.MarkAsReadRequest{NotificationId: 1},
			setupMock: func(mockRepo *repoMocks.MockNotificationRepository) {
				mockRepo.On("MarkAsRead", int64(1)).Return(assert.AnError)
			},
			expectedResult: &pb.MarkAsReadResponse{Success: false},
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.MockNotificationRepository)
			tt.setupMock(mockRepo)

			logger := logging.NewEmptyLogger()
			service := New(mockRepo, logger)
			result, err := service.MarkAsRead(context.Background(), tt.input)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult.Success, result.Success)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationService_MarkAllAsRead(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.MarkAllAsReadRequest
		setupMock      func(*repoMocks.MockNotificationRepository)
		expectedResult *pb.MarkAllAsReadResponse
		expectedError  bool
	}{
		{
			name:  "success - marks all notifications as read",
			input: &pb.MarkAllAsReadRequest{UserId: 123},
			setupMock: func(mockRepo *repoMocks.MockNotificationRepository) {
				mockRepo.On("MarkAllAsRead", int64(123)).Return(nil)
			},
			expectedResult: &pb.MarkAllAsReadResponse{Success: true},
			expectedError:  false,
		},
		{
			name:  "error - repository returns error",
			input: &pb.MarkAllAsReadRequest{UserId: 123},
			setupMock: func(mockRepo *repoMocks.MockNotificationRepository) {
				mockRepo.On("MarkAllAsRead", int64(123)).Return(assert.AnError)
			},
			expectedResult: &pb.MarkAllAsReadResponse{Success: false},
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.MockNotificationRepository)
			tt.setupMock(mockRepo)

			logger := logging.NewEmptyLogger()
			service := New(mockRepo, logger)
			result, err := service.MarkAllAsRead(context.Background(), tt.input)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult.Success, result.Success)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestToProtoNotificationResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    models.Notification
		expected *pb.NotificationResponse
	}{
		{
			name: "converts notification to proto response",
			input: models.Notification{
				NotificationID: 1,
				UserID:         123,
				Content:        "Test notification",
				IsRead:         false,
			},
			expected: &pb.NotificationResponse{
				NotificationId: 1,
				UserId:         123,
				Content:        "Test notification",
				IsRead:         false,
			},
		},
		{
			name: "handles zero values",
			input: models.Notification{
				NotificationID: 0,
				UserID:         0,
				Content:        "",
				IsRead:         false,
			},
			expected: &pb.NotificationResponse{
				NotificationId: 0,
				UserId:         0,
				Content:        "",
				IsRead:         false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoNotificationResponse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
