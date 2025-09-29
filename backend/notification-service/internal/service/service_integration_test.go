package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

var (
	testService pb.NotificationServiceServer
	testRepo    repository.NotificationRepository
)

type mockStream struct {
	notifications []*pb.NotificationResponse
	grpc.ServerStream
}

func (m *mockStream) Send(notification *pb.NotificationResponse) error {
	m.notifications = append(m.notifications, notification)
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

	var repoErr error
	testRepo, repoErr = repository.NewNotificationPgRepository()
	if repoErr != nil {
		fmt.Printf("Failed to create repository: %v\n", repoErr)
		os.Exit(1)
	}

	testService = service.New(testRepo)

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

func TestIntegrationNotificationService_GetByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	user1ID := insertTestUser(t, "user1")
	user2ID := insertTestUser(t, "user2")

	_, err := testRepo.Create(models.NotificationRequest{
		UserID:  user1ID,
		Content: "Test notification 1",
	})
	require.NoError(t, err)

	_, err = testRepo.Create(models.NotificationRequest{
		UserID:  user1ID,
		Content: "Test notification 2",
	})
	require.NoError(t, err)

	_, err = testRepo.Create(models.NotificationRequest{
		UserID:  user2ID,
		Content: "Other user notification",
	})
	require.NoError(t, err)

	req := &pb.GetByUserIDRequest{UserId: user1ID}
	stream := &mockStream{}

	err = testService.GetByUserID(req, stream)
	require.NoError(t, err)
	assert.Len(t, stream.notifications, 2)

	for _, notification := range stream.notifications {
		assert.Equal(t, user1ID, notification.UserId)
	}
}

func TestIntegrationNotificationService_MarkAsRead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	userID := insertTestUser(t, "user1")

	notification, err := testRepo.Create(models.NotificationRequest{
		UserID:  userID,
		Content: "Test notification",
	})
	require.NoError(t, err)

	initialNotification, err := testRepo.GetByID(notification.NotificationID)
	require.NoError(t, err)
	assert.False(t, initialNotification.IsRead)

	req := &pb.MarkAsReadRequest{NotificationId: notification.NotificationID}
	resp, err := testService.MarkAsRead(context.Background(), req)

	require.NoError(t, err)
	assert.True(t, resp.Success)

	updatedNotification, err := testRepo.GetByID(notification.NotificationID)
	require.NoError(t, err)
	assert.True(t, updatedNotification.IsRead)
}

func TestIntegrationNotificationService_MarkAllAsRead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)
	user1ID := insertTestUser(t, "user1")
	user2ID := insertTestUser(t, "user2")

	_, err := testRepo.Create(models.NotificationRequest{
		UserID:  user1ID,
		Content: "Test notification 1",
	})
	require.NoError(t, err)

	_, err = testRepo.Create(models.NotificationRequest{
		UserID:  user1ID,
		Content: "Test notification 2",
	})
	require.NoError(t, err)

	_, err = testRepo.Create(models.NotificationRequest{
		UserID:  user2ID,
		Content: "Other user notification",
	})
	require.NoError(t, err)

	req := &pb.MarkAllAsReadRequest{UserId: user1ID}
	resp, err := testService.MarkAllAsRead(context.Background(), req)

	require.NoError(t, err)
	assert.True(t, resp.Success)

	notifications, err := testRepo.GetByUserID(user1ID)
	require.NoError(t, err)

	for _, notification := range notifications {
		assert.True(t, notification.IsRead)
	}

	user2Notifications, err := testRepo.GetByUserID(user2ID)
	require.NoError(t, err)
	assert.False(t, user2Notifications[0].IsRead)
}

func cleanupTables(t *testing.T) {
	t.Helper()

	repo := testRepo.(*repository.NotificationPgRepository)
	_, err := repo.DB().Exec(context.Background(), "DELETE FROM notifications")
	if err != nil {
		t.Logf("Error cleaning up notifications: %v", err)
	}

	_, err = repo.DB().Exec(context.Background(), "DELETE FROM users")
	if err != nil {
		t.Logf("Error cleaning up users: %v", err)
	}
}

func insertTestUser(t *testing.T, username string) int64 {
	t.Helper()

	validBcryptHash := "$2a$10$abcdefghijklmnopqrstuuxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	repo := testRepo.(*repository.NotificationPgRepository)

	var userID int64
	err := repo.DB().QueryRow(context.Background(),
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING user_id",
		username, validBcryptHash).Scan(&userID)

	if err != nil {
		err = repo.DB().QueryRow(context.Background(),
			"SELECT user_id FROM users WHERE username = $1", username).Scan(&userID)
		require.NoError(t, err, "Failed to get user ID for username: %s", username)
	}

	return userID
}
