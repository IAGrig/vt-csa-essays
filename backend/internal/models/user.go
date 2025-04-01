package models

import "time"

// Internal DB representation
type User struct {
	ID           int       `json:"id" binding:"required,number"`
	Username     string    `json:"username" binding:"required,alphanum"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Login and registration DTO
type UserLoginRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,alphanum"`
}

// Public response without sensitive fields
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}
