package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/shared/pg_util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewPgRepository struct {
	db *pgxpool.Pool
}

func NewReviewPgRepository() (ReviewRepository, error) {
	pool, err := pgutil.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &ReviewPgRepository{db: pool}, nil
}

func (repository *ReviewPgRepository) Add(request models.ReviewRequest) (models.Review, error) {
	var r models.Review
	err := repository.db.QueryRow(context.Background(),
		`INSERT INTO reviews (essay_id, rank, content, author)
		VALUES ($1, $2, $3, $4)
		RETURNING review_id, created_at;`,
		request.EssayId,
		request.Rank,
		request.Content,
		request.Author,
	).Scan(&r.ID, &r.CreatedAt)

	if err != nil {
		return models.Review{}, fmt.Errorf("failed to create review: %w", err)
	}

	r.EssayId = request.EssayId
	r.Rank = request.Rank
	r.Content = request.Content
	r.Author = request.Author

	return r, nil
}

func (repository *ReviewPgRepository) GetAllReviews() ([]models.Review, error) {
	rows, err := repository.db.Query(context.Background(),
		`SELECT review_id, essay_id, rank, content, author, created_at
		FROM reviews
		ORDER BY created_at DESC;`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load reviews: %w", err)
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
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

func (repository *ReviewPgRepository) GetByEssayId(id int) ([]models.Review, error) {
	rows, err := repository.db.Query(context.Background(),
		`SELECT review_id, essay_id, rank, content, author, created_at
		FROM reviews
		WHERE essay_id = $1;`,
		id)
	if err != nil {
		return nil, fmt.Errorf("failed to load reviews: %w", err)
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
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

func (repository *ReviewPgRepository) RemoveById(id int) (models.Review, error) {
	var r models.Review
	err := repository.db.QueryRow(context.Background(),
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
			return models.Review{}, ReviewNotFoundErr
		}
		return models.Review{}, fmt.Errorf("failed to delete review: %w", err)
	}

	return r, nil
}

func (repository *ReviewPgRepository) DB() *pgxpool.Pool {
	return repository.db
}
