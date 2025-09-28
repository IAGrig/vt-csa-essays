package models

import (
	"time"

	reviewPb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

// Domain model
type Essay struct {
	ID        int
	Content   string
	Author    string
	CreatedAt time.Time
}

// Short get response
type EssayResponse struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// Detailed get response
type EssayWithReviewsResponse struct {
	ID        int                     `json:"id"`
	Content   string                  `json:"content"`
	Author    string                  `json:"author"`
	CreatedAt time.Time               `json:"created_at"`
	Reviews   []reviewPb.ReviewResponse `json:"reviews"`
}

// Add/update request DTO
type EssayRequest struct {
	Content string `json:"content" binding:"required"`
	Author  string `json:"author" binding:"required"`
}
