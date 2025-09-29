package repository

import (
	"errors"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/models"
)

var (
	NotificationNotFoundErr = errors.New("notification not found")
)

type NotificationRepository interface {
	Create(notification models.NotificationRequest) (models.Notification, error)
	GetByUserID(userID int64) ([]models.Notification, error)
	MarkAsRead(notificationID int64) error
	MarkAllAsRead(userID int64) error
	GetByID(notificationID int64) (models.Notification, error)
}
