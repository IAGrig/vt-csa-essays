package converters

import (
	"github.com/gin-gonic/gin"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

func MarshalNotificationResponse(n *pb.NotificationResponse) gin.H {
	if n == nil {
		return gin.H{}
	}
	return gin.H{
		"notification_id": n.NotificationId,
		"user_id":         n.UserId,
		"content":         n.Content,
		"is_read":         n.IsRead,
		"created_at":      n.CreatedAt,
	}
}
