package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/IAGrig/vt-csa-essays/backend/shared/pg_util"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EssayPgRepository struct {
	db     *pgxpool.Pool
	logger *logging.Logger
}

func NewEssayPgRepository(logger *logging.Logger) (*EssayPgRepository, error) {
	pool, err := pgutil.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &EssayPgRepository{db: pool, logger: logger}, nil
}

func (repository *EssayPgRepository) Add(request models.EssayRequest) (models.Essay, error) {
	logger := repository.logger.With(
		zap.String("operation", "add_essay"),
		zap.String("author", request.Author),
	)

	logger.Debug("Creating new essay")

	e, err := repository.GetByAuthorName(request.Author)
	if err == nil { // if essay found successfully
		logger.Warn("Duplicate essay creation attempt")
		return e, DuplicateErr
	}

	err = repository.db.QueryRow(context.Background(),
		`INSERT INTO essays (content, author)
		VALUES ($1, $2)
		RETURNING essay_id, content, author,
				(SELECT user_id FROM users WHERE username = $2) AS author_id,
				created_at;`,
		request.Content,
		request.Author,
	).Scan(&e.ID, &e.Content, &e.Author, &e.AuthorId, &e.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			logger.Warn("Database constraint violation - duplicate essay")
			return models.Essay{}, DuplicateErr
		}
		logger.Error("Failed to create essay in database", zap.Error(err))
		return models.Essay{}, fmt.Errorf("failed to create essay: %w", err)
	}

	logger.Info("Essay created successfully", zap.Int64("essay_id", int64(e.ID)))
	return e, nil
}

func (repository *EssayPgRepository) GetAllEssays() ([]models.Essay, error) {
	logger := repository.logger.With(zap.String("operation", "get_all_essays"))

	logger.Debug("Getting all essays")

	rows, err := repository.db.Query(context.Background(),
		`SELECT e.essay_id, e.content, e.author, u.user_id AS author_id, e.created_at
		FROM essays e
		JOIN users u ON e.author = u.username
		ORDER BY created_at DESC;`,
	)
	if err != nil {
		logger.Error("Failed to get essays from database", zap.Error(err))
		return nil, fmt.Errorf("failed to get essays: %w", err)
	}
	defer rows.Close()

	var essays []models.Essay
	for rows.Next() {
		var e models.Essay
		rows.Scan(
			&e.ID,
			&e.Content,
			&e.Author,
			&e.AuthorId,
			&e.CreatedAt,
		)
		essays = append(essays, e)
	}

	logger.Debug("Retrieved essays", zap.Int("count", len(essays)))
	return essays, nil
}

func (repository *EssayPgRepository) GetByAuthorName(username string) (models.Essay, error) {
	logger := repository.logger.With(
		zap.String("operation", "get_essay_by_author"),
		zap.String("author", username),
	)

	logger.Debug("Getting essay by author name")

	var e models.Essay
	err := repository.db.QueryRow(context.Background(),
		`SELECT e.essay_id, e.content, e.author, u.user_id AS author_id, e.created_at
		FROM essays e
		JOIN users u ON e.author = u.username
		WHERE author = $1;`,
		username,
	).Scan(&e.ID, &e.Content, &e.Author, &e.AuthorId, &e.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("Essay not found")
			return models.Essay{}, EssayNotFoundErr
		}
		logger.Error("Failed to get essay from database", zap.Error(err))
		return models.Essay{}, fmt.Errorf("failed to get essay: %w", err)
	}

	logger.Debug("Essay retrieved successfully", zap.Int64("essay_id", int64(e.ID)))
	return e, nil
}

func (repository *EssayPgRepository) RemoveByAuthorName(username string) (models.Essay, error) {
	logger := repository.logger.With(
		zap.String("operation", "remove_essay_by_author"),
		zap.String("author", username),
	)

	logger.Info("Removing essay by author name")

	tx, err := repository.db.Begin(context.Background())
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return models.Essay{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	var e models.Essay
	err = tx.QueryRow(context.Background(),
		`DELETE FROM essays
		WHERE author = $1
		RETURNING essay_id, content, author,
				(SELECT user_id FROM users WHERE username = $1) AS author_id,
				created_at;`,
		username,
	).Scan(&e.ID, &e.Content, &e.Author, &e.AuthorId, &e.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn("Essay not found for deletion")
			return models.Essay{}, EssayNotFoundErr
		}
		logger.Error("Failed to delete essay from database", zap.Error(err))
		return models.Essay{}, fmt.Errorf("failed to delete essay: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return models.Essay{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Essay deleted successfully", zap.Int64("essay_id", int64(e.ID)))
	return e, nil
}

func (repository *EssayPgRepository) SearchByContent(content string) ([]models.Essay, error) {
	logger := repository.logger.With(
		zap.String("operation", "search_essays_by_content"),
		zap.String("search_term", content),
	)

	logger.Debug("Searching essays by content")

	rows, err := repository.db.Query(context.Background(),
		`SELECT e.essay_id, e.content, e.author, u.user_id AS author_id, e.created_at, similarity(lower(content), lower($1)) as siml
		FROM essays e
		JOIN users u ON e.author = u.username
		ORDER BY siml DESC
		LIMIT 20;`,
		content)
	if err != nil {
		logger.Error("Failed to search essays in database", zap.Error(err))
		return nil, fmt.Errorf("failed to get essays: %w", err)
	}
	defer rows.Close()

	var essays []models.Essay
	for rows.Next() {
		var e models.Essay
		rows.Scan(
			&e.ID,
			&e.Content,
			&e.Author,
			&e.AuthorId,
			&e.CreatedAt,
			nil,
		)
		essays = append(essays, e)
	}

	logger.Debug("Search completed", zap.Int("results_count", len(essays)))
	return essays, nil
}

func (repository *EssayPgRepository) DB() *pgxpool.Pool {
	return repository.db
}
