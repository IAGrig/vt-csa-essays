package mocks

import (
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(notification models.NotificationRequest) (models.Notification, error) {
	args := m.Called(notification)
	return args.Get(0).(models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetByUserID(userID int64) ([]models.Notification, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsRead(notificationID int64) error {
	args := m.Called(notificationID)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAllAsRead(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetByID(notificationID int64) (models.Notification, error) {
	args := m.Called(notificationID)
	return args.Get(0).(models.Notification), args.Error(1)
}
