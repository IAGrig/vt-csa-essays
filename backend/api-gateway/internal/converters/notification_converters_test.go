package converters

import (
	"testing"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMarshalNotificationResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.NotificationResponse
		expected gin.H
	}{
		{
			name: "success - converts complete notification response",
			input: &pb.NotificationResponse{
				NotificationId: 1,
				UserId:         123,
				Content:        "Your essay has been reviewed!",
				IsRead:         false,
				CreatedAt:      1234567890,
			},
			expected: gin.H{
				"notification_id": int64(1),
				"user_id":         int64(123),
				"content":         "Your essay has been reviewed!",
				"is_read":         false,
				"created_at":      int64(1234567890),
			},
		},
		{
			name: "success - converts notification with zero values",
			input: &pb.NotificationResponse{
				NotificationId: 0,
				UserId:         0,
				Content:        "",
				IsRead:         false,
				CreatedAt:      0,
			},
			expected: gin.H{
				"notification_id": int64(0),
				"user_id":         int64(0),
				"content":         "",
				"is_read":         false,
				"created_at":      int64(0),
			},
		},
		{
			name:     "success - handles nil input",
			input:    nil,
			expected: gin.H{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MarshalNotificationResponse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
