package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
	reviewPb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

var (
	testService pb.EssayServiceServer
	testRepo    *repository.EssayPgRepository
)

type mockReviewClient struct {
	reviewsByEssayId map[int32][]*reviewPb.ReviewResponse
}

func (m *mockReviewClient) Add(ctx context.Context, in *reviewPb.ReviewAddRequest, opts ...grpc.CallOption) (*reviewPb.ReviewResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockReviewClient) GetAllReviews(ctx context.Context, in *reviewPb.EmptyRequest, opts ...grpc.CallOption) (reviewPb.ReviewService_GetAllReviewsClient, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockReviewClient) GetByEssayId(ctx context.Context, in *reviewPb.GetByEssayIdRequest, opts ...grpc.CallOption) (reviewPb.ReviewService_GetByEssayIdClient, error) {
	reviews, exists := m.reviewsByEssayId[in.EssayId]
	if !exists {
		reviews = []*reviewPb.ReviewResponse{}
	}
	return &mockReviewStream{reviews: reviews}, nil
}

func (m *mockReviewClient) RemoveById(ctx context.Context, in *reviewPb.RemoveByIdRequest, opts ...grpc.CallOption) (*reviewPb.ReviewResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

type mockReviewStream struct {
	reviews []*reviewPb.ReviewResponse
	index   int
	grpc.ClientStream
}

func (m *mockReviewStream) Recv() (*reviewPb.ReviewResponse, error) {
	if m.index >= len(m.reviews) {
		return nil, context.Canceled
	}
	review := m.reviews[m.index]
	m.index++
	return review, nil
}

func (m *mockReviewStream) Context() context.Context {
	return context.Background()
}

func (m *mockReviewStream) Header() (metadata.MD, error) {
	return nil, nil
}

func (m *mockReviewStream) Trailer() metadata.MD {
	return nil
}

func (m *mockReviewStream) CloseSend() error {
	return nil
}

type mockEssayStream struct {
	essays []*pb.EssayResponse
	grpc.ServerStream
}

func (m *mockEssayStream) Send(essay *pb.EssayResponse) error {
	m.essays = append(m.essays, essay)
	return nil
}

func (m *mockEssayStream) Context() context.Context {
	return context.Background()
}

func (m *mockEssayStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockEssayStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockEssayStream) SetTrailer(md metadata.MD) {
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

	var repoErr error
	testRepo, repoErr = repository.NewEssayPgRepository()
	if repoErr != nil {
		fmt.Printf("Failed to create repository: %v\n", repoErr)
		os.Exit(1)
	}

	mockReviewClient := &mockReviewClient{
		reviewsByEssayId: make(map[int32][]*reviewPb.ReviewResponse),
	}

	testService = service.New(testRepo, mockReviewClient)

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

func TestIntegrationEssayService_Add(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "test-author")

	req := &pb.EssayAddRequest{
		Content: "This is a test essay content",
		Author:  "test-author",
	}

	ctx := context.Background()
	resp, err := testService.Add(ctx, req)

	require.NoError(t, err)
	assert.NotZero(t, resp.Id)
	assert.Equal(t, req.Content, resp.Content)
	assert.Equal(t, req.Author, resp.Author)
	assert.InDelta(t, time.Now().Unix(), resp.CreatedAt, 2*time.Second.Seconds())
}

func TestIntegrationEssayService_GetAllEssays(t *testing.T) {
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

	req := &pb.EmptyRequest{}
	stream := &mockEssayStream{}

	err = testService.GetAllEssays(req, stream)
	require.NoError(t, err)
	assert.Len(t, stream.essays, 2)
}

func TestIntegrationEssayService_RemoveByAuthorName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "test-author")

	essayReq := models.EssayRequest{Content: "Test essay content", Author: "test-author"}
	addedEssay, err := testRepo.Add(essayReq)
	require.NoError(t, err)

	req := &pb.RemoveByAuthorNameRequest{
		Authorname: "test-author",
	}

	ctx := context.Background()
	resp, err := testService.RemoveByAuthorName(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int32(addedEssay.ID), resp.Id)
	assert.Equal(t, essayReq.Content, resp.Content)
	assert.Equal(t, essayReq.Author, resp.Author)

	_, err = testRepo.GetByAuthorName("test-author")
	assert.ErrorIs(t, err, repository.EssayNotFoundErr)
}

func TestIntegrationEssayService_SearchByContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	insertTestUser(t, "author1")
	insertTestUser(t, "author2")

	essay1 := models.EssayRequest{Content: "This essay talks about artificial intelligence", Author: "author1"}
	essay2 := models.EssayRequest{Content: "This essay discusses machine learning", Author: "author2"}

	_, err := testRepo.Add(essay1)
	require.NoError(t, err)
	_, err = testRepo.Add(essay2)
	require.NoError(t, err)

	req := &pb.SearchByContentRequest{
		Content: "artificial intelligence",
	}

	stream := &mockEssayStream{}

	err = testService.SearchByContent(req, stream)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(stream.essays), 1)

	found := false
	for _, essay := range stream.essays {
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
