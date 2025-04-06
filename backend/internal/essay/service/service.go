package service

import (
	"github.com/IAGrig/vt-csa-essays/internal/essay"
	essaystore "github.com/IAGrig/vt-csa-essays/internal/essay/store"
	reviewstore "github.com/IAGrig/vt-csa-essays/internal/review/store"
)

type EssaySevice interface {
	Add(essay.EssayRequest) (essay.EssayResponse, error)
	GetByAuthorName(string) (essay.EssayWithReviewsResponse, error)
	RemoveByAuthorName(string) (essay.EssayResponse, error)
}

type service struct {
	essayStore  essaystore.EssayStore
	reviewStore reviewstore.ReviewStore
}

func New(essayStore essaystore.EssayStore, reviewStore reviewstore.ReviewStore) EssaySevice {
	return &service{essayStore: essayStore, reviewStore: reviewStore}
}

func (service *service) Add(request essay.EssayRequest) (essay.EssayResponse, error) {
	e, err := service.essayStore.Add(request)
	if err != nil {
		return essay.EssayResponse{}, nil
	}

	return service.essayToResponse(e)
}

func (service *service) GetByAuthorName(authorname string) (essay.EssayWithReviewsResponse, error) {
	e, err := service.essayStore.GetByAuthorName(authorname)
	if err != nil {
		return essay.EssayWithReviewsResponse{}, nil
	}

	return service.essayToResponseWithReviews(e)
}

func (service *service) RemoveByAuthorName(authorname string) (essay.EssayResponse, error) {
	e, err := service.essayStore.RemoveByAuthorName(authorname)
	if err != nil {
		return essay.EssayResponse{}, nil
	}

	return service.essayToResponse(e)
}
