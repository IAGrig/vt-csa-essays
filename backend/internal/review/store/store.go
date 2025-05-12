package store

import "github.com/IAGrig/vt-csa-essays/internal/review"

type ReviewStore interface {
	Add(review review.ReviewRequest) (review.Review, error)
	GetAllReviews() ([]review.Review, error)
	GetByEssayId(id int) ([]review.Review, error)
	RemoveById(id int) (review.Review, error)
}
