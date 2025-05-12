package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/internal/essay"
	"github.com/IAGrig/vt-csa-essays/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EssayPgStore struct {
	db *pgxpool.Pool
}

func NewEssayPgStore() (*EssayPgStore, error) {
	pool, err := util.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &EssayPgStore{db: pool}, nil
}

func (store *EssayPgStore) Add(request essay.EssayRequest) (essay.Essay, error) {
	var e essay.Essay
	err := store.db.QueryRow(context.Background(),
		`INSERT INTO essays (content, author)
		VALUES ($1, $2)
		RETURNING essay_id, content, author, created_at;`,
		request.Content,
		request.Author,
	).Scan(&e.ID, &e.Content, &e.Author, &e.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return essay.Essay{}, DuplicateErr
		}
		return essay.Essay{}, fmt.Errorf("failed to create essay: %w", err)
	}

	return e, nil
}

func (store *EssayPgStore) GetAllEssays() ([]essay.Essay, error) {
	rows, err := store.db.Query(context.Background(),
		`SELECT essay_id, content, author, created_at
		FROM essays
		ORDER BY created_at DESC;`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get essays: %w", err)
	}
	defer rows.Close()

	var essays []essay.Essay
	for rows.Next() {
		var e essay.Essay
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

func (store *EssayPgStore) GetByAuthorName(username string) (essay.Essay, error) {
	var e essay.Essay
	err := store.db.QueryRow(context.Background(),
		`SELECT essay_id, content, author, created_at
		FROM essays
		WHERE author = $1;`,
		username,
	).Scan(&e.ID, &e.Content, &e.Author, &e.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return essay.Essay{}, EssayNotFoundErr
		}
		return essay.Essay{}, fmt.Errorf("failed to get essay: %w", err)
	}

	return e, nil
}

func (store *EssayPgStore) RemoveByAuthorName(username string) (essay.Essay, error) {
	tx, err := store.db.Begin(context.Background())
	if err != nil {
		return essay.Essay{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	var e essay.Essay
	err = tx.QueryRow(context.Background(),
		`DELETE FROM essays
		WHERE author = $1
		RETURNING essay_id, content, author, created_at;`,
		username,
	).Scan(&e.ID, &e.Content, &e.Author, &e.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return essay.Essay{}, EssayNotFoundErr
		}
		return essay.Essay{}, fmt.Errorf("failed to delete essay: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return essay.Essay{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return e, nil
}
