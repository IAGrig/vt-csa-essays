package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/IAGrig/vt-csa-essays/backend/shared/pg_util"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPgRepository struct {
	db     *pgxpool.Pool
	logger *logging.Logger
}

func NewUserPgRepository(logger *logging.Logger) (UserRepository, error) {
	pool, err := pgutil.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &UserPgRepository{db: pool, logger: logger}, nil
}

func (repository *UserPgRepository) Add(request models.UserLoginRequest) (models.User, error) {
	logger := repository.logger.With(
		zap.String("operation", "add_user"),
		zap.String("username", request.Username),
	)

	logger.Debug("Creating new user")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	var user models.User
	err = repository.db.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING user_id, username, created_at;`,
		request.Username, passwordHash).Scan(&user.ID, &user.Username, &user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			logger.Warn("Duplicate user creation attempt")
			return models.User{}, DuplicateErr
		}
		logger.Error("Failed to create user in database", zap.Error(err))
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("User created successfully", zap.Int64("user_id", int64(user.ID)))
	return user, nil
}

func (repository *UserPgRepository) Auth(request models.UserLoginRequest) (models.User, error) {
	logger := repository.logger.With(
		zap.String("operation", "auth_user"),
		zap.String("username", request.Username),
	)

	logger.Debug("Authenticating user")

	var user models.User

	err := repository.db.QueryRow(
		context.Background(),
		`SELECT user_id, username, password_hash, created_at
		FROM users
		WHERE username = $1;`,
		request.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn("User not found during authentication")
			return models.User{}, NotFoundErr
		}
		logger.Error("Database error during authentication", zap.Error(err))
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		logger.Warn("Invalid password provided")
		return models.User{}, AuthErr
	}

	response := models.User{ID: user.ID, Username: user.Username, CreatedAt: user.CreatedAt}
	logger.Debug("User authenticated successfully", zap.Int64("user_id", int64(user.ID)))
	return response, nil
}

func (repository *UserPgRepository) GetByUsername(username string) (models.User, error) {
	logger := repository.logger.With(
		zap.String("operation", "get_user_by_username"),
		zap.String("username", username),
	)

	logger.Debug("Getting user by username")

	var user models.User

	err := repository.db.QueryRow(context.Background(),
		`SELECT user_id, username, created_at
		FROM users
		WHERE username = $1;`,
		username).Scan(&user.ID, &user.Username, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn("User not found")
			return models.User{}, NotFoundErr
		}
		logger.Error("Database error when getting user", zap.Error(err))
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	logger.Debug("User retrieved successfully", zap.Int64("user_id", int64(user.ID)))
	return user, nil
}

func (repository *UserPgRepository) DB() *pgxpool.Pool {
	return repository.db
}
