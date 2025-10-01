package repository

import (
	"errors"
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/models"
)

var (
	AuthErr      = errors.New("authorization failed")
	DuplicateErr = errors.New("user already exists")
	NotFoundErr  = errors.New("user not found")
)

type UserRepository interface {
	Add(user models.UserLoginRequest) (models.User, error)
	Auth(request models.UserLoginRequest) (models.User, error)
	GetByUsername(username string) (models.User, error)
}
