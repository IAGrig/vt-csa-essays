package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/IAGrig/vt-csa-essays/backend/shared/pg_util"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewPgRepository struct {
	db     *pgxpool.Pool
	logger *logging.Logger
}

func NewReviewPgRepository(logger *logging.Logger) (ReviewRepository, error) {
	pool, err := pgutil.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &ReviewPgRepository{db: pool, logger: logger}, nil
}

func (repository *ReviewPgRepository) Add(request models.ReviewRequest) (models.Review, error) {
	logger := repository.logger.With(
		zap.String("operation", "add_review"),
		zap.Int("essay_id", request.EssayId),
		zap.String("author", request.Author),
	)

	logger.Debug("Creating new review")

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
		logger.Error("Failed to create review in database", zap.Error(err))
		return models.Review{}, fmt.Errorf("failed to create review: %w", err)
	}

	r.EssayId = request.EssayId
	r.Rank = request.Rank
	r.Content = request.Content
	r.Author = request.Author

	logger.Info("Review created successfully",
		zap.Int("review_id", r.ID))
	return r, nil
}

func (repository *ReviewPgRepository) GetAllReviews() ([]models.Review, error) {
	logger := repository.logger.With(zap.String("operation", "get_all_reviews"))

	logger.Debug("Getting all reviews")

	rows, err := repository.db.Query(context.Background(),
		`SELECT review_id, essay_id, rank, content, author, created_at
		FROM reviews
		ORDER BY created_at DESC;`,
	)
	if err != nil {
		logger.Error("Failed to get reviews from database", zap.Error(err))
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
			logger.Error("Failed to scan review row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		reviews = append(reviews, r)
	}

	logger.Debug("Retrieved reviews", zap.Int("count", len(reviews)))
	return reviews, nil
}

func (repository *ReviewPgRepository) GetByEssayId(id int) ([]models.Review, error) {
	logger := repository.logger.With(
		zap.String("operation", "get_reviews_by_essay_id"),
		zap.Int("essay_id", id),
	)

	logger.Debug("Getting reviews by essay ID")

	rows, err := repository.db.Query(context.Background(),
		`SELECT review_id, essay_id, rank, content, author, created_at
		FROM reviews
		WHERE essay_id = $1;`,
		id)
	if err != nil {
		logger.Error("Failed to get reviews by essay ID from database", zap.Error(err))
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
			logger.Error("Failed to scan review row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		reviews = append(reviews, r)
	}

	logger.Debug("Retrieved reviews for essay", zap.Int("count", len(reviews)))
	return reviews, nil
}

func (repository *ReviewPgRepository) RemoveById(id int) (models.Review, error) {
	logger := repository.logger.With(
		zap.String("operation", "remove_review_by_id"),
		zap.Int("review_id", id),
	)

	logger.Debug("Removing review by ID")

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
			logger.Debug("Review not found for removal")
			return models.Review{}, ReviewNotFoundErr
		}
		logger.Error("Failed to delete review from database", zap.Error(err))
		return models.Review{}, fmt.Errorf("failed to delete review: %w", err)
	}

	logger.Info("Review removed successfully")
	return r, nil
}

func (repository *ReviewPgRepository) DB() *pgxpool.Pool {
	return repository.db
}
