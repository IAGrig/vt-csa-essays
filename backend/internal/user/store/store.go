package store

import "github.com/IAGrig/vt-csa-essays/internal/user"

type UserStore interface {
	// Creates a new user with the provided login request details
	// Returns the created user response
	Add(user user.UserLoginRequest) (user.UserResponse, error)

	// Authenticates a user based on the provided login request
	// Returns the authenticated user response or an error
	Auth(request user.UserLoginRequest) (user.UserResponse, error)

	GetByUsername(username string) (user.UserResponse, error)
}
