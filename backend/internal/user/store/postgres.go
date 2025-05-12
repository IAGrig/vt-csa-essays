package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/internal/user"
	"github.com/IAGrig/vt-csa-essays/internal/util"
	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPgStore struct {
	db *pgxpool.Pool
}

func NewUserPgStore() (*UserPgStore, error) {
	pool, err := util.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &UserPgStore{db: pool}, nil
}

func (store *UserPgStore) Add(request user.UserLoginRequest) (user.UserResponse, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		return user.UserResponse{}, fmt.Errorf("failed to hash password: %w", err)
	}

	var usr user.UserResponse
	err = store.db.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING user_id, username, created_at;`,
		request.Username, passwordHash).Scan(&usr.ID, &usr.Username, &usr.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return user.UserResponse{}, DuplicateErr
		}
		return user.UserResponse{}, fmt.Errorf("failed to create user: %w", err)
	}

	return usr, nil
}

func (store *UserPgStore) Auth(request user.UserLoginRequest) (user.UserResponse, error) {
	var usr user.User

	err := store.db.QueryRow(
		context.Background(),
		`SELECT user_id, username, password_hash, created_at
		FROM users
		WHERE username = $1;`,
		request.Username,
	).Scan(&usr.ID, &usr.Username, &usr.PasswordHash, &usr.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.UserResponse{}, NotFoundErr
		}
		return user.UserResponse{}, fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(request.Password)); err != nil {
		return user.UserResponse{}, AuthErr
	}

	response := user.UserResponse{ID: usr.ID, Username: usr.Username, CreatedAt: usr.CreatedAt}
	return response, nil
}

func (store *UserPgStore) GetByUsername(username string) (user.UserResponse, error) {
	var usr user.UserResponse

	err := store.db.QueryRow(context.Background(),
		`SELECT user_id, username, created_at
		FROM users
		WHERE username = $1;`,
		username).Scan(&usr.ID, &usr.Username, &usr.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.UserResponse{}, NotFoundErr
		}
		return user.UserResponse{}, fmt.Errorf("failed to get user: %w", err)
	}

	return usr, nil
}
