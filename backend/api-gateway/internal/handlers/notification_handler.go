package handlers

import (
	"net/http"
	"strconv"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/converters"
	"github.com/gin-gonic/gin"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

type NotificationHandler struct {
	notificationClient clients.NotificationClient
}

func NewNotificationHandler(notificationClient clients.NotificationClient) *NotificationHandler {
	return &NotificationHandler{notificationClient: notificationClient}
}

// GET /api/notifications
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.notificationClient.GetByUserID(
		c.Request.Context(),
		&pb.GetByUserIDRequest{UserId: userID.(int64)},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var notifications []gin.H
	for _, notification := range resp {
		notifications = append(notifications, converters.MarshalNotificationResponse(notification))
	}

	c.JSON(http.StatusOK, notifications)
}

// POST /api/notifications/:notificationId/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	notificationIdStr := c.Param("notificationId")
	notificationId, err := strconv.ParseInt(notificationIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	resp, err := h.notificationClient.MarkAsRead(
		c.Request.Context(),
		&pb.MarkAsReadRequest{NotificationId: notificationId},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// POST /api/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.notificationClient.MarkAllAsRead(
		c.Request.Context(),
		&pb.MarkAllAsReadRequest{UserId: userID.(int64)},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
