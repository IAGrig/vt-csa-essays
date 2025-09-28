package mocks

import (
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Add(user models.UserLoginRequest) (models.User, error) {
	args := m.Called(user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) Auth(request models.UserLoginRequest) (models.User, error) {
	args := m.Called(request)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (models.User, error) {
	args := m.Called(username)
	return args.Get(0).(models.User), args.Error(1)
}
