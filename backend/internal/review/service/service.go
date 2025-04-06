package service

import (
	"github.com/IAGrig/vt-csa-essays/internal/review"
	"github.com/IAGrig/vt-csa-essays/internal/review/store"
)

type ReviewService interface {
	Add(review.ReviewRequest) (review.ReviewResponse, error)
	GetByEssayId(int) ([]review.ReviewResponse, error)
	RemoveById(int) (review.ReviewResponse, error)
}

type service struct {
	store store.ReviewStore
}

func New(store store.ReviewStore) ReviewService {
	return &service{store: store}
}

func (service *service) Add(request review.ReviewRequest) (review.ReviewResponse, error) {
	r, err := service.store.Add(request)
	if err != nil {
		return review.ReviewResponse{}, err
	}

	return service.ReviewToResponse(r)
}

func (service *service) GetByEssayId(id int) ([]review.ReviewResponse, error) {
	reviews, err := service.store.GetByEssayId(id)
	if err != nil {
		return []review.ReviewResponse{}, err
	}

	responses := make([]review.ReviewResponse, len(reviews))
	for i, r := range reviews {
		response, err := service.ReviewToResponse(r)
		if err != nil {
			return []review.ReviewResponse{}, err
		}

		responses[i] = response
	}

	return responses, nil
}

func (service *service) RemoveById(id int) (review.ReviewResponse, error) {
	r, err := service.store.RemoveById(id)
	if err != nil {
		return review.ReviewResponse{}, err
	}

	return service.ReviewToResponse(r)
}
