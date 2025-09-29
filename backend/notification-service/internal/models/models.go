package models

import "time"

// Domain model
type Notification struct {
	NotificationID int64     `db:"notification_id"`
	UserID         int64     `db:"user_id"`
	Content        string    `db:"content"`
	IsRead         bool      `db:"is_read"`
	CreatedAt      time.Time `db:"created_at"`
}

// Get response
type NotificationResponse struct {
	NotificationID int64     `json:"notification_id"`
	UserID         int64     `json:"user_id"`
	Content        string    `json:"content"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

// Create request DTO
type NotificationRequest struct {
	UserID  int64  `json:"user_id" binding:"required,number"`
	Content string `json:"content" binding:"required"`
}
