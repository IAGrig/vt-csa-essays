package service

import "github.com/IAGrig/vt-csa-essays/internal/review"

func (service *service) ReviewToResponse(r review.Review) (review.ReviewResponse, error) {
	return review.ReviewResponse{
		ID:        r.ID,
		EssayId:   r.EssayId,
		Rank:      r.Rank,
		Content:   r.Content,
		Author:    r.Author,
		CreatedAt: r.CreatedAt,
	}, nil
}
