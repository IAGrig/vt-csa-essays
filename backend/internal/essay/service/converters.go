package service

import "github.com/IAGrig/vt-csa-essays/internal/essay"

func (service *service) essayToResponse(e essay.Essay) (essay.EssayResponse, error) {
	return essay.EssayResponse{
		ID:        e.ID,
		Content:   e.Content,
		Author:    e.Author,
		CreatedAt: e.CreatedAt,
	}, nil
}

func (service *service) essayToResponseWithReviews(e essay.Essay) (essay.EssayWithReviewsResponse, error) {
	reviews, err := service.reviewService.GetByEssayId(e.ID)
	if err != nil {
		return essay.EssayWithReviewsResponse{}, err
	}

	return essay.EssayWithReviewsResponse{
		ID:        e.ID,
		Content:   e.Content,
		Author:    e.Author,
		CreatedAt: e.CreatedAt,
		Reviews:   reviews,
	}, nil
}
