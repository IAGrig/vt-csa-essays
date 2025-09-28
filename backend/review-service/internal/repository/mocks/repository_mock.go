package mocks

import (
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockReviewRepository struct {
	mock.Mock
}

func (m *MockReviewRepository) Add(review models.ReviewRequest) (models.Review, error) {
	args := m.Called(review)
	return args.Get(0).(models.Review), args.Error(1)
}

func (m *MockReviewRepository) GetAllReviews() ([]models.Review, error) {
	args := m.Called()
	return args.Get(0).([]models.Review), args.Error(1)
}

func (m *MockReviewRepository) GetByEssayId(id int) ([]models.Review, error) {
	args := m.Called(id)
	return args.Get(0).([]models.Review), args.Error(1)
}

func (m *MockReviewRepository) RemoveById(id int) (models.Review, error) {
	args := m.Called(id)
	return args.Get(0).(models.Review), args.Error(1)
}
