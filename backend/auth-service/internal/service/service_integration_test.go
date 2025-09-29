package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/service"
	"github.com/IAGrig/vt-csa-essays/backend/shared/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)

var (
	testService pb.UserServiceServer
	testRepo    repository.UserRepository
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

	accessSecret := []byte("test-access-sectet")
	refresSecret := []byte("test-refresh-sectet")
	jwtGenerator := jwt.NewGenerator(accessSecret, refresSecret)
	jwtParser := jwt.NewParser(accessSecret, refresSecret)

	testService = service.New(testRepo, jwtGenerator, jwtParser)

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

func TestIntegrationAuthService_Register(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	req := &pb.UserRegisterRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()
	resp, err := testService.Register(ctx, req)

	require.NoError(t, err)
	assert.NotZero(t, resp.Id)
	assert.Equal(t, req.Username, resp.Username)
	assert.InDelta(t, time.Now().Unix(), resp.CreatedAt, 2*time.Second.Seconds())
}

func TestIntegrationAuthService_Register_Duplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	req := &pb.UserRegisterRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()

	_, err := testService.Register(ctx, req)
	require.NoError(t, err)

	_, err = testService.Register(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), repository.DuplicateErr.Error())
}

func TestIntegrationAuthService_Auth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	registerReq := &pb.UserRegisterRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()
	_, err := testService.Register(ctx, registerReq)
	require.NoError(t, err)

	authReq := &pb.UserLoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	resp, err := testService.Auth(ctx, authReq)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestIntegrationAuthService_Auth_WrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	registerReq := &pb.UserRegisterRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()
	_, err := testService.Register(ctx, registerReq)
	require.NoError(t, err)

	authReq := &pb.UserLoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	_, err = testService.Auth(ctx, authReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), repository.AuthErr.Error())
}

func TestIntegrationAuthService_GetByUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	registerReq := &pb.UserRegisterRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()
	registeredUser, err := testService.Register(ctx, registerReq)
	require.NoError(t, err)

	getReq := &pb.GetByUsernameRequest{
		Username: "testuser",
	}

	user, err := testService.GetByUsername(ctx, getReq)
	require.NoError(t, err)
	assert.Equal(t, registeredUser.Id, user.Id)
	assert.Equal(t, registeredUser.Username, user.Username)
	assert.Equal(t, registeredUser.CreatedAt, user.CreatedAt)
}

func TestIntegrationAuthService_RefreshToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanupTables(t)

	registerReq := &pb.UserRegisterRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()
	_, err := testService.Register(ctx, registerReq)
	require.NoError(t, err)

	authReq := &pb.UserLoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	authResp, err := testService.Auth(ctx, authReq)
	require.NoError(t, err)

	refreshReq := &pb.RefreshTokenRequest{
		RefreshToken: authResp.RefreshToken,
	}

	refreshResp, err := testService.RefreshToken(ctx, refreshReq)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshResp.AccessToken)
	assert.Empty(t, refreshResp.RefreshToken)
}

func cleanupTables(t *testing.T) {
	t.Helper()

	repo := testRepo.(*repository.UserPgRepository)
	_, err := repo.DB().Exec(context.Background(), "DELETE FROM users")
	require.NoError(t, err)
}
