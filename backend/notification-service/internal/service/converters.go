package service

import (
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/models"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

func toProtoNotificationResponse(notification models.Notification) *pb.NotificationResponse {
	var createdAt int64
	if !notification.CreatedAt.IsZero() {
		createdAt = notification.CreatedAt.Unix()
	}

	return &pb.NotificationResponse{
		NotificationId: notification.NotificationID,
		UserId:         notification.UserID,
		Content:        notification.Content,
		IsRead:         notification.IsRead,
		CreatedAt:      createdAt,
	}
}
