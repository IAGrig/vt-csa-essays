package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/internal/review"
	"github.com/IAGrig/vt-csa-essays/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewPgStore struct {
	db *pgxpool.Pool
}

func NewReviewPgStore() (ReviewStore, error) {
	pool, err := util.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &ReviewPgStore{db: pool}, nil
}

func (store *ReviewPgStore) Add(request review.ReviewRequest) (review.Review, error) {
	var r review.Review
	err := store.db.QueryRow(context.Background(),
		`INSERT INTO reviews (essay_id, rank, content, author)
		VALUES ($1, $2, $3, $4)
		RETURNING review_id, created_at;`,
		request.EssayId,
		request.Rank,
		request.Content,
		request.Author,
	).Scan(&r.ID, &r.CreatedAt)

	if err != nil {
		return review.Review{}, fmt.Errorf("failed to create review: %w", err)
	}

	r.EssayId = request.EssayId
	r.Rank = request.Rank
	r.Content = request.Content
	r.Author = request.Author

	return r, nil
}

func (store *ReviewPgStore) GetByEssayId(id int) ([]review.Review, error) {
	rows, err := store.db.Query(context.Background(),
		`SELECT review_id, essay_id, rank, content, author, created_at
		FROM reviews
		WHERE essay_id = $1;`,
		id)
	if err != nil {
		return nil, fmt.Errorf("failed to load reviews: %w", err)
	}
	defer rows.Close()

	var reviews []review.Review
	for rows.Next() {
		var r review.Review
		err = rows.Scan(
			&r.ID,
			&r.EssayId,
			&r.Rank,
			&r.Content,
			&r.Author,
			&r.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

func (store *ReviewPgStore) RemoveById(id int) (review.Review, error) {
	var r review.Review
	err := store.db.QueryRow(context.Background(),
		`DELETE FROM reviews
			WHERE review_id = $1
			RETURNING review_id, essay_id, rank, content, author, created_at;`,
		id,
	).Scan(
		&r.ID,
		&r.EssayId,
		&r.Rank,
		&r.Content,
		&r.Author,
		&r.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return review.Review{}, ReviewNotFoundErr
		}
		return review.Review{}, fmt.Errorf("failed to delete review: %w", err)
	}

	return r, nil
}
