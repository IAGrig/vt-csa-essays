package db

import "github.com/IAGrig/vt-csa-essays/internal/models"

type UserStore interface {
	Add(user models.UserLoginRequest) (models.UserResponse, error)
	Get(username string) (models.UserResponse, error)
}
