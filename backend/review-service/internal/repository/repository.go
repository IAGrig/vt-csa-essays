package repository

import (
	"errors"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
)

var (
	ReviewNotFoundErr = errors.New("review not found")
)

type ReviewRepository interface {
	Add(review models.ReviewRequest) (models.Review, error)
	GetAllReviews() ([]models.Review, error)
	GetByEssayId(id int) ([]models.Review, error)
	RemoveById(id int) (models.Review, error)
}
