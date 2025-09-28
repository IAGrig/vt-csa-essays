package mocks

import (
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockEssayRepository struct {
	mock.Mock
}

func (m *MockEssayRepository) Add(essay models.EssayRequest) (models.Essay, error) {
	args := m.Called(essay)
	return args.Get(0).(models.Essay), args.Error(1)
}

func (m *MockEssayRepository) GetAllEssays() ([]models.Essay, error) {
	args := m.Called()
	return args.Get(0).([]models.Essay), args.Error(1)
}

func (m *MockEssayRepository) GetByAuthorName(username string) (models.Essay, error) {
	args := m.Called(username)
	return args.Get(0).(models.Essay), args.Error(1)
}

func (m *MockEssayRepository) RemoveByAuthorName(username string) (models.Essay, error) {
	args := m.Called(username)
	return args.Get(0).(models.Essay), args.Error(1)
}

func (m *MockEssayRepository) SearchByContent(query string) ([]models.Essay, error) {
	args := m.Called(query)
	return args.Get(0).([]models.Essay), args.Error(1)
}
