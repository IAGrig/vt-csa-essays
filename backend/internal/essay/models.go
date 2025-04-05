package essay

import "time"

// Domain model
type Essay struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// Add/update request DTO
type EssayRequest struct {
	Content string `json:"content" binding:"required"`
	Author  string `json:"author" binding:"required"`
}
