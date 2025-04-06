package store

import (
	"errors"
	"time"

	"slices"

	"github.com/IAGrig/vt-csa-essays/internal/review"
)

var (
	EssayNotFoundErr  = errors.New("essay not found")
	ReviewNotFoundErr = errors.New("review not found")
)

// !Must be used only for test purposes!
type ReviewMemStore struct {
	reviewsByEssayId map[int][]review.Review
	reviewsCount     int
}

// Creates an instance of ReviewMemStore, !must be used only for test purposes!
func NewReviewMemStore() ReviewStore {
	reviewsByEssayId := make(map[int][]review.Review)

	return &ReviewMemStore{
		reviewsByEssayId: reviewsByEssayId,
		reviewsCount:     0,
	}
}

func (store *ReviewMemStore) Add(request review.ReviewRequest) (review.Review, error) {
	review := review.Review{
		ID:        store.reviewsCount + 1,
		EssayId:   request.EssayId,
		Rank:      request.Rank,
		Content:   request.Content,
		Author:    request.Author,
		CreatedAt: time.Now(),
	}

	store.reviewsByEssayId[request.EssayId] = append(store.reviewsByEssayId[request.EssayId], review)
	store.reviewsCount++
	return review, nil
}

func (store *ReviewMemStore) GetByEssayId(id int) ([]review.Review, error) {
	if reviews, ok := store.reviewsByEssayId[id]; ok {
		return reviews, nil
	}
	return []review.Review{}, nil
}

func (store *ReviewMemStore) RemoveById(id int) (review.Review, error) {
	essayId, index := -1, -1
	var r review.Review
	for eId, reviews := range store.reviewsByEssayId {
		for i, review := range reviews {
			if review.ID == id {
				essayId = eId
				index = i
				r = review
				break
			}
		}
		if essayId != -1 {
			break
		}
	}
	if essayId == -1 {
		return review.Review{}, ReviewNotFoundErr
	}

	slice := store.reviewsByEssayId[essayId]
	slice = slices.Delete(slice, index, index+1)

	return r, nil
}
