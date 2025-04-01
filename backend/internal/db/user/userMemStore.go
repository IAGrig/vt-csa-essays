package db

import (
	"errors"
	"time"

	"github.com/IAGrig/vt-csa-essays/internal/crypto"
	"github.com/IAGrig/vt-csa-essays/internal/models"
)

var (
	DublicateErr = errors.New("user already exists")
	NotFoundErr  = errors.New("not found")
)

type UserMemStore struct {
	list map[string]models.User
}

func NewUserMemStore() *UserMemStore {
	list := make(map[string]models.User)
	return &UserMemStore{
		list,
	}
}

func (store UserMemStore) Add(request models.UserLoginRequest) (models.UserResponse, error) {
	if _, err := store.Get(request.Username); err != NotFoundErr {
		return models.UserResponse{}, DublicateErr
	}

	passwordHash, err := crypto.GenerateHash([]byte(request.Password))
	if err != nil {
		return models.UserResponse{}, err
	}

	user := models.User{
		ID:           len(store.list) + 1,
		Username:     request.Username,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
	}

	store.list[request.Username] = user
	return store.userToUserResponse(user), nil
}

func (store UserMemStore) Get(name string) (models.UserResponse, error) {

	if val, ok := store.list[name]; ok {
		return store.userToUserResponse(val), nil
	}

	return models.UserResponse{}, NotFoundErr
}

func (store UserMemStore) userToUserResponse(user models.User) models.UserResponse {
	return models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}
}
