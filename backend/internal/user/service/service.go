package service

import (
	"github.com/IAGrig/vt-csa-essays/internal/auth/jwt"
	"github.com/IAGrig/vt-csa-essays/internal/user"
	"github.com/IAGrig/vt-csa-essays/internal/user/store"
)

type UserSevice interface {
	Add(user.UserLoginRequest) (user.UserResponse, error)
	Auth(user.UserLoginRequest) (accessToken, refreshToken string, err error)
	GetByUsername(string) (user.UserResponse, error)
	RefreshToken(string) (accessToken, refreshToken string, err error)
}

type service struct {
	store        store.UserStore
	jwtGenerator jwt.TokenGenerator
	jwtParser    jwt.TokenParser
}

func New(
	store store.UserStore,
	jwtGenerator jwt.TokenGenerator,
	jwtParser jwt.TokenParser,
) UserSevice {
	return &service{
		store:        store,
		jwtGenerator: jwtGenerator,
		jwtParser:    jwtParser,
	}
}

func (service *service) Add(request user.UserLoginRequest) (user.UserResponse, error) {
	return service.store.Add(request)
}

func (service *service) Auth(request user.UserLoginRequest) (string, string, error) {
	user, err := service.store.Auth(request)
	if err != nil {
		return "", "", err
	}

	accessToken, err := service.jwtGenerator.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := service.jwtGenerator.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (service *service) GetByUsername(username string) (user.UserResponse, error) {
	return service.store.GetByUsername(username)
}

func (service *service) RefreshToken(refreshToken string) (string, string, error) {
	username, err := service.jwtParser.GetUsername(refreshToken, "refresh")
	if err != nil {
		return "", "", err
	}

	user, err := service.GetByUsername(username)
	if err != nil {
		return "", "", err
	}

	newAccessToken, err := service.jwtGenerator.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// I need to implement tokens invalidation and regenerate refresh tokens on every request

	return newAccessToken, "", nil
}
