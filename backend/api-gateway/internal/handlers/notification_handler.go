package handlers

import (
	"net/http"
	"strconv"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/converters"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

type NotificationHandler struct {
	notificationClient clients.NotificationClient
	logger            *logging.Logger
}

func NewNotificationHandler(notificationClient clients.NotificationClient, logger *logging.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationClient: notificationClient,
		logger:            logger,
	}
}

// GET /api/notifications
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("Authentication required for getting notifications")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userIDInt := userID.(int64)
	logger := h.logger.With(
		zap.String("operation", "get_user_notifications"),
		zap.Int64("user_id", userIDInt),
	)

	logger.Debug("Get user notifications request")
	resp, err := h.notificationClient.GetByUserID(
		c.Request.Context(),
		&pb.GetByUserIDRequest{UserId: userIDInt},
	)
	if err != nil {
		logger.Error("Failed to get user notifications",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var notifications []gin.H
	for _, notification := range resp {
		notifications = append(notifications, converters.MarshalNotificationResponse(notification))
	}

	logger.Debug("Retrieved user notifications",
		zap.Int("count", len(notifications)))
	c.JSON(http.StatusOK, notifications)
}

// POST /api/notifications/:notificationId/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	notificationIdStr := c.Param("notificationId")
	notificationId, err := strconv.ParseInt(notificationIdStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid notification ID",
			zap.String("notification_id", notificationIdStr),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	logger := h.logger.With(
		zap.String("operation", "mark_notification_as_read"),
		zap.Int64("notification_id", notificationId),
	)

	logger.Debug("Mark notification as read request")
	resp, err := h.notificationClient.MarkAsRead(
		c.Request.Context(),
		&pb.MarkAsReadRequest{NotificationId: notificationId},
	)
	if err != nil {
		logger.Error("Failed to mark notification as read",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !resp.Success {
		logger.Warn("Notification not found for marking as read")
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	logger.Debug("Notification marked as read successfully")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// POST /api/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("Authentication required for marking all notifications as read")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userIDInt := userID.(int64)
	logger := h.logger.With(
		zap.String("operation", "mark_all_notifications_as_read"),
		zap.Int64("user_id", userIDInt),
	)

	logger.Debug("Mark all notifications as read request")
	resp, err := h.notificationClient.MarkAllAsRead(
		c.Request.Context(),
		&pb.MarkAllAsReadRequest{UserId: userIDInt},
	)
	if err != nil {
		logger.Error("Failed to mark all notifications as read",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !resp.Success {
		logger.Error("Failed to mark all notifications as read - service returned failure")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notifications as read"})
		return
	}

	logger.Debug("All notifications marked as read successfully")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
