package models

import "time"

// Domain model
type Review struct {
	ID        int
	EssayId   int
	Rank      int
	Content   string
	Author    string
	CreatedAt time.Time
}

// Get response
type ReviewResponse struct {
	ID        int       `json:"id"`
	EssayId   int       `json:"essayId"`
	Rank      int       `json:"rank"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// Add/update request DTO
type ReviewRequest struct {
	EssayId int    `json:"essayId" binding:"required,number"`
	Rank    int    `json:"rank" binding:"required,number,gte=1,lte=3"`
	Content string `json:"content" binding:"required"`
	Author  string `json:"author" binding:"required"`
}
