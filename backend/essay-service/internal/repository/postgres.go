package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/shared/pg_util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EssayPgRepository struct {
	db *pgxpool.Pool
}

func NewEssayPgRepository() (*EssayPgRepository, error) {
	pool, err := pgutil.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &EssayPgRepository{db: pool}, nil
}

func (repository *EssayPgRepository) Add(request models.EssayRequest) (models.Essay, error) {
	e, err := repository.GetByAuthorName(request.Author)
	if err == nil { // if essay found successfully
		return e, DuplicateErr
	}

	err = repository.db.QueryRow(context.Background(),
		`INSERT INTO essays (content, author)
		VALUES ($1, $2)
		RETURNING essay_id, content, author, created_at;`,
		request.Content,
		request.Author,
	).Scan(&e.ID, &e.Content, &e.Author, &e.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.Essay{}, DuplicateErr
		}
		return models.Essay{}, fmt.Errorf("failed to create essay: %w", err)
	}

	return e, nil
}

func (repository *EssayPgRepository) GetAllEssays() ([]models.Essay, error) {
	rows, err := repository.db.Query(context.Background(),
		`SELECT essay_id, content, author, created_at
		FROM essays
		ORDER BY created_at DESC;`,
	)
	if err != nil {
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
			&e.CreatedAt,
		)
		essays = append(essays, e)
	}

	return essays, nil
}

func (repository *EssayPgRepository) GetByAuthorName(username string) (models.Essay, error) {
	var e models.Essay
	err := repository.db.QueryRow(context.Background(),
		`SELECT essay_id, content, author, created_at
		FROM essays
		WHERE author = $1;`,
		username,
	).Scan(&e.ID, &e.Content, &e.Author, &e.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Essay{}, EssayNotFoundErr
		}
		return models.Essay{}, fmt.Errorf("failed to get essay: %w", err)
	}

	return e, nil
}

func (repository *EssayPgRepository) RemoveByAuthorName(username string) (models.Essay, error) {
	tx, err := repository.db.Begin(context.Background())
	if err != nil {
		return models.Essay{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	var e models.Essay
	err = tx.QueryRow(context.Background(),
		`DELETE FROM essays
		WHERE author = $1
		RETURNING essay_id, content, author, created_at;`,
		username,
	).Scan(&e.ID, &e.Content, &e.Author, &e.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Essay{}, EssayNotFoundErr
		}
		return models.Essay{}, fmt.Errorf("failed to delete essay: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return models.Essay{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return e, nil
}

func (repository *EssayPgRepository) SearchByContent(content string) ([]models.Essay, error) {
	rows, err := repository.db.Query(context.Background(),
		`SELECT essay_id, content, author, created_at, similarity(lower(content), lower($1)) as siml
		FROM essays
		ORDER BY siml DESC
		LIMIT 20;`,
		content)
	if err != nil {
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
			&e.CreatedAt,
			nil,
		)
		essays = append(essays, e)
	}

	return essays, nil
}

func (repository *EssayPgRepository) DB() *pgxpool.Pool {
	return repository.db
}
