package store

import (
	"errors"
	"time"

	"github.com/IAGrig/vt-csa-essays/internal/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	AuthErr      = errors.New("authorization failed")
	DuplicateErr = errors.New("user already exists")
	NotFoundErr  = errors.New("user not found")
)

type UserMemStore struct {
	list map[string]user.User
}

func NewUserMemStore() *UserMemStore {
	list := make(map[string]user.User)
	return &UserMemStore{
		list,
	}
}

func (store *UserMemStore) Add(request user.UserLoginRequest) (user.UserResponse, error) {
	if _, err := store.GetByUsername(request.Username); err != NotFoundErr {
		return user.UserResponse{}, DuplicateErr
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		return user.UserResponse{}, err
	}

	user := user.User{
		ID:           len(store.list) + 1,
		Username:     request.Username,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
	}

	store.list[request.Username] = user
	return store.userToUserResponse(user), nil
}

func (store *UserMemStore) Auth(request user.UserLoginRequest) (user.UserResponse, error) {
	usr, err := store.getUser(request.Username)
	if err != nil {
		return user.UserResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(request.Password)); err != nil {
		return user.UserResponse{}, AuthErr
	}

	return store.userToUserResponse(usr), nil
}

// Internal store func, returns a full-featured User object containing PasswordHash
func (store *UserMemStore) getUser(username string) (user.User, error) {
	if user, ok := store.list[username]; ok {
		return user, nil
	}
	return user.User{}, NotFoundErr
}

func (store *UserMemStore) GetByUsername(username string) (user.UserResponse, error) {
	user, err := store.getUser(username)
	return store.userToUserResponse(user), err
}

func (store *UserMemStore) userToUserResponse(usr user.User) user.UserResponse {
	return user.UserResponse{
		ID:        usr.ID,
		Username:  usr.Username,
		CreatedAt: usr.CreatedAt,
	}
}
