package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/shared/pg_util"
	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPgRepository struct {
	db *pgxpool.Pool
}

func NewUserPgRepository() (UserRepository, error) {
	pool, err := pgutil.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &UserPgRepository{db: pool}, nil
}

func (repository *UserPgRepository) Add(request models.UserLoginRequest) (models.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	var usr models.User
	err = repository.db.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING user_id, username, created_at;`,
		request.Username, passwordHash).Scan(&usr.ID, &usr.Username, &usr.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, DuplicateErr
		}
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return usr, nil
}

func (repository *UserPgRepository) Auth(request models.UserLoginRequest) (models.User, error) {
	var usr models.User

	err := repository.db.QueryRow(
		context.Background(),
		`SELECT user_id, username, password_hash, created_at
		FROM users
		WHERE username = $1;`,
		request.Username,
	).Scan(&usr.ID, &usr.Username, &usr.PasswordHash, &usr.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, NotFoundErr
		}
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(request.Password)); err != nil {
		return models.User{}, AuthErr
	}

	response := models.User{ID: usr.ID, Username: usr.Username, CreatedAt: usr.CreatedAt}
	return response, nil
}

func (repository *UserPgRepository) GetByUsername(username string) (models.User, error) {
	var usr models.User

	err := repository.db.QueryRow(context.Background(),
		`SELECT user_id, username, created_at
		FROM users
		WHERE username = $1;`,
		username).Scan(&usr.ID, &usr.Username, &usr.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, NotFoundErr
		}
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return usr, nil
}

func (repository *UserPgRepository) DB() *pgxpool.Pool {
	return repository.db
}
