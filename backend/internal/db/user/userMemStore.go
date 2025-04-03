package db

import (
	"errors"
	"time"

	"github.com/IAGrig/vt-csa-essays/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	AuthErr      = errors.New("authorization failed")
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

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
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

func (store UserMemStore) Auth(request models.UserLoginRequest) (models.UserResponse, error) {
	user, err := store.getUser(request.Username)
	if err != nil {
		return models.UserResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return models.UserResponse{}, AuthErr
	}

	return store.userToUserResponse(user), nil
}

// Internal store func, returns a full-featured User object containing PasswordHash
func (store UserMemStore) getUser(username string) (models.User, error) {
	if user, ok := store.list[username]; ok {
		return user, nil
	}
	return models.User{}, NotFoundErr
}

func (store UserMemStore) Get(username string) (models.UserResponse, error) {
	user, err := store.getUser(username)
	return store.userToUserResponse(user), err
}

func (store UserMemStore) userToUserResponse(user models.User) models.UserResponse {
	return models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}
}
