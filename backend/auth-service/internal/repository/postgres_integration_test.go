package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	testRepo repository.UserRepository
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := setupTestDB(ctx)
	if err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if container != nil {
			_ = container.Terminate(ctx)
		}
	}()

	var repoErr error
	testRepo, repoErr = repository.NewUserPgRepository()
	if repoErr != nil {
		fmt.Printf("Failed to create repository: %v\n", repoErr)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func setupTestDB(ctx context.Context) (testcontainers.Container, error) {
	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return container, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return container, fmt.Errorf("failed to get container port: %w", err)
	}

	os.Setenv("POSTGRES_HOST", host)
	os.Setenv("POSTGRES_PORT", port.Port())
	os.Setenv("POSTGRES_USER", "test_user")
	os.Setenv("POSTGRES_PASSWORD", "test_password")
	os.Setenv("POSTGRES_DB_NAME", "test_db")
	os.Setenv("POSTGRES_SSL_MODE", "disable")

	fmt.Printf("Database running at: %s:%s\n", host, port.Port())

	if err := runGooseMigrations(host, port.Port()); err != nil {
		return container, fmt.Errorf("failed to run goose migrations: %w", err)
	}

	return container, nil
}

func runGooseMigrations(host, port string) error {
	connStr := fmt.Sprintf("postgres://test_user:test_password@%s:%s/test_db?sslmode=disable", host, port)

	migrationsPath, err := getMigrationsPath()
	if err != nil {
		return err
	}

	fmt.Printf("Running goose migrations from: %s\n", migrationsPath)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		return fmt.Errorf("failed to run goose up: %w", err)
	}

	fmt.Println("Goose migrations completed successfully")
	return nil
}

func getMigrationsPath() (string, error) {
	testDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	possiblePaths := []string{
		filepath.Join(testDir, "..", "..", "..", "migrations"),
		filepath.Join(testDir, "..", "..", "migrations"),
	}

	for _, path := range possiblePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("migrations directory not found. Tried: %v", possiblePaths)
}

func TestIntegrationUserRepository_Add(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	userReq := models.UserLoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	user, err := testRepo.Add(userReq)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, userReq.Username, user.Username)
	assert.False(t, user.CreatedAt.IsZero())
}

func TestIntegrationUserRepository_Add_Duplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	userReq := models.UserLoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	_, err := testRepo.Add(userReq)
	require.NoError(t, err)

	_, err = testRepo.Add(userReq)
	assert.ErrorIs(t, err, repository.DuplicateErr)
}

func TestIntegrationUserRepository_Auth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	userReq := models.UserLoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	_, err := testRepo.Add(userReq)
	require.NoError(t, err)

	user, err := testRepo.Auth(userReq)
	require.NoError(t, err)
	assert.Equal(t, userReq.Username, user.Username)
	assert.NotZero(t, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
}

func TestIntegrationUserRepository_Auth_WrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	userReq := models.UserLoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	_, err := testRepo.Add(userReq)
	require.NoError(t, err)

	wrongReq := models.UserLoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	_, err = testRepo.Auth(wrongReq)
	assert.ErrorIs(t, err, repository.AuthErr)
}

func TestIntegrationUserRepository_Auth_UserNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	userReq := models.UserLoginRequest{
		Username: "nonexistent",
		Password: "password",
	}

	_, err := testRepo.Auth(userReq)
	assert.ErrorIs(t, err, repository.NotFoundErr)
}

func TestIntegrationUserRepository_GetByUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	userReq := models.UserLoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	addedUser, err := testRepo.Add(userReq)
	require.NoError(t, err)

	user, err := testRepo.GetByUsername(userReq.Username)
	require.NoError(t, err)
	assert.Equal(t, addedUser.ID, user.ID)
	assert.Equal(t, addedUser.Username, user.Username)
	assert.Equal(t, addedUser.CreatedAt, user.CreatedAt)
}

func TestIntegrationUserRepository_GetByUsername_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	_, err := testRepo.GetByUsername("nonexistent")
	assert.ErrorIs(t, err, repository.NotFoundErr)
}

func cleanupTables(t *testing.T) {
	t.Helper()
	repo := testRepo.(*repository.UserPgRepository)
	_, err := repo.DB().Exec(context.Background(), "DELETE FROM users")
	require.NoError(t, err)
}
