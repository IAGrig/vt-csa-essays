package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	testRepo *repository.EssayPgRepository
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

	logger := logging.NewEmptyLogger()
	var repoErr error
	testRepo, repoErr = repository.NewEssayPgRepository(logger)
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

func TestIntegrationEssayRepository_Add(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "test-author")

	essayReq := models.EssayRequest{
		Content: "This is a test essay content",
		Author:  "test-author",
	}

	essay, err := testRepo.Add(essayReq)
	require.NoError(t, err)
	assert.NotZero(t, essay.ID)
	assert.Equal(t, essayReq.Content, essay.Content)
	assert.Equal(t, essayReq.Author, essay.Author)
	assert.NotZero(t, essay.AuthorId)
	assert.False(t, essay.CreatedAt.IsZero())
}

func TestIntegrationEssayRepository_Add_Duplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "test-author")

	essayReq := models.EssayRequest{
		Content: "This is a test essay content",
		Author:  "test-author",
	}

	_, err := testRepo.Add(essayReq)
	require.NoError(t, err)

	_, err = testRepo.Add(essayReq)
	assert.ErrorIs(t, err, repository.DuplicateErr)
}

func TestIntegrationEssayRepository_GetAllEssays(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "author1")
	insertTestUser(t, "author2")

	essay1 := models.EssayRequest{Content: "Essay 1 content", Author: "author1"}
	essay2 := models.EssayRequest{Content: "Essay 2 content", Author: "author2"}

	_, err := testRepo.Add(essay1)
	require.NoError(t, err)
	_, err = testRepo.Add(essay2)
	require.NoError(t, err)

	essays, err := testRepo.GetAllEssays()
	require.NoError(t, err)
	assert.Len(t, essays, 2)
}

func TestIntegrationEssayRepository_GetByAuthorName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "test-author")

	essayReq := models.EssayRequest{
		Content: "This is a test essay content",
		Author:  "test-author",
	}

	addedEssay, err := testRepo.Add(essayReq)
	require.NoError(t, err)

	essay, err := testRepo.GetByAuthorName("test-author")
	require.NoError(t, err)
	assert.Equal(t, addedEssay.ID, essay.ID)
	assert.Equal(t, addedEssay.Content, essay.Content)
	assert.Equal(t, addedEssay.Author, essay.Author)
	assert.NotZero(t, essay.AuthorId)
}

func TestIntegrationEssayRepository_GetByAuthorName_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	_, err := testRepo.GetByAuthorName("nonexistent")
	assert.ErrorIs(t, err, repository.EssayNotFoundErr)
}

func TestIntegrationEssayRepository_RemoveByAuthorName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "test-author")

	essayReq := models.EssayRequest{
		Content: "This is a test essay content",
		Author:  "test-author",
	}

	addedEssay, err := testRepo.Add(essayReq)
	require.NoError(t, err)

	removedEssay, err := testRepo.RemoveByAuthorName("test-author")
	require.NoError(t, err)
	assert.Equal(t, addedEssay.ID, removedEssay.ID)
	assert.Equal(t, addedEssay.Content, removedEssay.Content)
	assert.Equal(t, addedEssay.Author, removedEssay.Author)
	assert.Equal(t, addedEssay.AuthorId, addedEssay.AuthorId)

	_, err = testRepo.GetByAuthorName("test-author")
	assert.ErrorIs(t, err, repository.EssayNotFoundErr)
}

func TestEssayIntegrationRepository_RemoveByAuthorName_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	_, err := testRepo.RemoveByAuthorName("nonexistent")
	assert.ErrorIs(t, err, repository.EssayNotFoundErr)
}

func TestIntegrationEssayRepository_SearchByContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "author1")
	insertTestUser(t, "author2")

	essay1 := models.EssayRequest{Content: "This essay talks about artificial intelligence", Author: "author1"}
	essay2 := models.EssayRequest{Content: "This essay discusses machine learning algorithms", Author: "author2"}

	_, err := testRepo.Add(essay1)
	require.NoError(t, err)
	_, err = testRepo.Add(essay2)
	require.NoError(t, err)

	essays, err := testRepo.SearchByContent("artificial intelligence")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(essays), 1)

	found := false
	for _, essay := range essays {
		if essay.Content == "This essay talks about artificial intelligence" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find essay with matching content")
}

func cleanupTables(t *testing.T) {
	t.Helper()
	_, err := testRepo.DB().Exec(context.Background(), `
		DELETE FROM essays;
		DELETE FROM users;
	`)
	require.NoError(t, err)
}

func insertTestUser(t *testing.T, username string) {
	t.Helper()

	validBcryptHash := "$2a$10$abcdefghijklmnopqrstuuxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	_, err := testRepo.DB().Exec(context.Background(),
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) ON CONFLICT (username) DO NOTHING",
		username, validBcryptHash)
	require.NoError(t, err)
}
