package mocks

import "github.com/stretchr/testify/mock"

type MockTokenParser struct {
	mock.Mock
}

func (m *MockTokenParser) GetUsername(token string, tokenType string) (string, error) {
	args := m.Called(token, tokenType)
	return args.String(0), args.Error(1)
}

func (m *MockTokenParser) GetUserId(token string, tokenType string) (int, error) {
	args := m.Called(token, tokenType)
	return args.Int(0), args.Error(1)
}
