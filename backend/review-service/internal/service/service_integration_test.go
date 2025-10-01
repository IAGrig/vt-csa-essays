package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/service"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

var (
	testService pb.ReviewServiceServer
	testRepo    repository.ReviewRepository
)

type mockStream struct {
	reviews []*pb.ReviewResponse
	grpc.ServerStream
}

func (m *mockStream) Send(review *pb.ReviewResponse) error {
	m.reviews = append(m.reviews, review)
	return nil
}

func (m *mockStream) Context() context.Context {
	return context.Background()
}

func (m *mockStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockStream) SetTrailer(md metadata.MD) {
}

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
	testRepo, repoErr = repository.NewReviewPgRepository(logger)
	if repoErr != nil {
		fmt.Printf("Failed to create repository: %v\n", repoErr)
		os.Exit(1)
	}

	testService = service.New(testRepo, nil, logger)

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

func TestIntegrationReviewService_Add(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "test-author")
	insertTestEssay(t, 1, "test-author")

	req := &pb.ReviewAddRequest{
		EssayId:       1,
		EssayAuthorId: 1,
		Rank:          2,
		Content:       "Test review content",
		Author:        "test-author",
	}

	ctx := context.Background()
	resp, err := testService.Add(ctx, req)

	require.NoError(t, err)
	assert.NotZero(t, resp.Id)
	assert.Equal(t, req.EssayId, resp.EssayId)
	assert.Equal(t, req.Rank, resp.Rank)
	assert.Equal(t, req.Content, resp.Content)
	assert.Equal(t, req.Author, resp.Author)
	assert.NotZero(t, resp.CreatedAt)
}

func TestIntegrationReviewService_GetByEssayId(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "author1")
	insertTestUser(t, "author2")
	insertTestUser(t, "test-author")
	insertTestEssay(t, 1, "test-author")
	insertTestEssay(t, 2, "test-author")

	_, err := testRepo.Add(models.ReviewRequest{EssayId: 1, Rank: 2, Content: "Review 1", Author: "author1"})
	require.NoError(t, err)
	_, err = testRepo.Add(models.ReviewRequest{EssayId: 1, Rank: 1, Content: "Review 2", Author: "author2"})
	require.NoError(t, err)
	_, err = testRepo.Add(models.ReviewRequest{EssayId: 2, Rank: 3, Content: "Review 3", Author: "author1"})
	require.NoError(t, err)

	req := &pb.GetByEssayIdRequest{EssayId: 1}
	stream := &mockStream{}

	err = testService.GetByEssayId(req, stream)
	require.NoError(t, err)
	assert.Len(t, stream.reviews, 2)

	for _, review := range stream.reviews {
		assert.Equal(t, int32(1), review.EssayId)
	}
}

func TestIntegrationReviewService_GetAllReviews(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "author1")
	insertTestUser(t, "author2")
	insertTestUser(t, "test-author")
	insertTestEssay(t, 1, "test-author")
	insertTestEssay(t, 2, "test-author")

	_, err := testRepo.Add(models.ReviewRequest{EssayId: 1, Rank: 2, Content: "Review 1", Author: "author1"})
	require.NoError(t, err)
	_, err = testRepo.Add(models.ReviewRequest{EssayId: 2, Rank: 1, Content: "Review 2", Author: "author2"})
	require.NoError(t, err)

	req := &pb.EmptyRequest{}
	stream := &mockStream{}

	err = testService.GetAllReviews(req, stream)
	require.NoError(t, err)
	assert.Len(t, stream.reviews, 2)
}

func TestIntegrationReviewService_RemoveById(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "test-author")
	insertTestEssay(t, 1, "test-author")

	addedReview, err := testRepo.Add(models.ReviewRequest{
		EssayId: 1,
		Rank:    2,
		Content: "To be removed",
		Author:  "test-author",
	})
	require.NoError(t, err)

	ctx := context.Background()
	req := &pb.RemoveByIdRequest{Id: int32(addedReview.ID)}
	resp, err := testService.RemoveById(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int32(addedReview.ID), resp.Id)

	stream := &mockStream{}
	err = testService.GetByEssayId(&pb.GetByEssayIdRequest{EssayId: 1}, stream)
	require.NoError(t, err)
	assert.Len(t, stream.reviews, 0)
}

func cleanupTables(t *testing.T) {
	t.Helper()

	reviews, err := testRepo.GetAllReviews()
	if err != nil {
		return
	}

	for _, review := range reviews {
		_, _ = testRepo.RemoveById(review.ID)
	}
}

func insertTestUser(t *testing.T, username string) {
	t.Helper()

	validBcryptHash := "$2a$10$abcdefghijklmnopqrstuuxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	repo := testRepo.(*repository.ReviewPgRepository)
	_, err := repo.DB().Exec(context.Background(),
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) ON CONFLICT (username) DO NOTHING",
		username, validBcryptHash)
	require.NoError(t, err)
}

func insertTestEssay(t *testing.T, essayId int, author string) {
	t.Helper()
	repo := testRepo.(*repository.ReviewPgRepository)
	_, err := repo.DB().Exec(context.Background(),
		"INSERT INTO essays (essay_id, content, author) VALUES ($1, $2, $3) ON CONFLICT (essay_id) DO NOTHING",
		essayId, "Test essay content", author)
	require.NoError(t, err)
}
