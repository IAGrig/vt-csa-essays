package db

import "github.com/IAGrig/vt-csa-essays/internal/models"

type UserStore interface {
	// Creates a new user with the provided login request details
	// Returns the created user response
	Add(user models.UserLoginRequest) (models.UserResponse, error)

	// Authenticates a user based on the provided login request
	// Returns the authenticated user response or an error
	Auth(request models.UserLoginRequest) (models.UserResponse, error)

	Get(username string) (models.UserResponse, error)
}
