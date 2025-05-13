package service

import (
	"github.com/IAGrig/vt-csa-essays/internal/essay"
	"github.com/IAGrig/vt-csa-essays/internal/essay/store"
	reviewservice "github.com/IAGrig/vt-csa-essays/internal/review/service"
)

type EssaySevice interface {
	Add(essay.EssayRequest) (essay.EssayResponse, error)
	GetAllEssays() ([]essay.EssayResponse, error)
	GetByAuthorName(string) (essay.EssayWithReviewsResponse, error)
	RemoveByAuthorName(string) (essay.EssayResponse, error)
	SearchByContent(string) ([]essay.EssayResponse, error)
}

type service struct {
	essayStore    store.EssayStore
	reviewService reviewservice.ReviewService
}

func New(essayStore store.EssayStore, reviewService reviewservice.ReviewService) EssaySevice {
	return &service{essayStore: essayStore, reviewService: reviewService}
}

func (service *service) Add(request essay.EssayRequest) (essay.EssayResponse, error) {
	e, err := service.essayStore.Add(request)
	if err != nil {
		return essay.EssayResponse{}, err
	}

	return service.essayToResponse(e)
}

func (service *service) GetAllEssays() ([]essay.EssayResponse, error) {
	essays, err := service.essayStore.GetAllEssays()
	if err != nil {
		return nil, err
	}

	responses, err := service.essaysToResponces(essays)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

func (service *service) GetByAuthorName(authorname string) (essay.EssayWithReviewsResponse, error) {
	e, err := service.essayStore.GetByAuthorName(authorname)
	if err != nil {
		return essay.EssayWithReviewsResponse{}, err
	}

	return service.essayToResponseWithReviews(e)
}

func (service *service) RemoveByAuthorName(authorname string) (essay.EssayResponse, error) {
	e, err := service.essayStore.RemoveByAuthorName(authorname)
	if err != nil {
		return essay.EssayResponse{}, err
	}

	return service.essayToResponse(e)
}

func (service *service) SearchByContent(content string) ([]essay.EssayResponse, error) {
	essays, err := service.essayStore.SearchByContent(content)
	if err != nil {
		return nil, err
	}

	responses, err := service.essaysToResponces(essays)
	if err != nil {
		return nil, err
	}

	return responses, nil
}
