package mocks

import (
	"github.com/IAGrig/vt-csa-essays/backend/shared/jwt"
	"github.com/stretchr/testify/mock"
)

type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) GenerateAccessToken(userInfo jwt.UserInfo) (string, error) {
	args := m.Called(userInfo)
	return args.String(0), args.Error(1)
}

func (m *MockTokenGenerator) GenerateRefreshToken(userInfo jwt.UserInfo) (string, error) {
	args := m.Called(userInfo)
	return args.String(0), args.Error(1)
}
