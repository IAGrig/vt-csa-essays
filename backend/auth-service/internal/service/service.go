package service

import (
	"context"

	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/shared/jwt"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"go.uber.org/zap"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)

type authService struct {
	pb.UnimplementedUserServiceServer
	repository   repository.UserRepository
	jwtGenerator jwt.TokenGenerator
	jwtParser    jwt.TokenParser
	logger       *logging.Logger
}

func New(
	repository repository.UserRepository,
	jwtGenerator jwt.TokenGenerator,
	jwtParser jwt.TokenParser,
	logger *logging.Logger,
) pb.UserServiceServer {
	return &authService{
		repository:   repository,
		jwtGenerator: jwtGenerator,
		jwtParser:    jwtParser,
		logger:       logger,
	}
}

func (s *authService) Register(ctx context.Context, in *pb.UserRegisterRequest) (*pb.UserResponse, error) {
	logger := s.logger.With(zap.String("operation", "register"), zap.String("username", in.Username))

	logger.Info("User registration request")

	req := models.UserLoginRequest{Username: in.Username, Password: in.Password}
	user, err := s.repository.Add(req)
	if err != nil {
		logger.Error("User registration failed", zap.Error(err))
		return nil, err
	}

	logger.Info("User registered successfully", zap.Int64("user_id", int64(user.ID)))
	return toProtoUserResponse(user), nil
}

func (s *authService) Auth(ctx context.Context, in *pb.UserLoginRequest) (*pb.AuthTokensResponse, error) {
	logger := s.logger.With(zap.String("operation", "auth"), zap.String("username", in.Username))

	logger.Debug("User authentication request")

	req := models.UserLoginRequest{Username: in.Username, Password: in.Password}

	user, err := s.repository.Auth(req)
	if err != nil {
		logger.Warn("User authentication failed", zap.Error(err))
		return nil, err
	}

	userInfo := jwt.UserInfo{UserId: user.ID, Username: user.Username}
	accessToken, err := s.jwtGenerator.GenerateAccessToken(userInfo)
	if err != nil {
		logger.Error("Failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := s.jwtGenerator.GenerateRefreshToken(userInfo)
	if err != nil {
		logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	logger.Info("User authenticated successfully", zap.Int64("user_id", int64(user.ID)))
	resp := &pb.AuthTokensResponse{AccessToken: accessToken, RefreshToken: refreshToken}

	return resp, nil
}

func (s *authService) GetByUsername(ctx context.Context, in *pb.GetByUsernameRequest) (*pb.UserResponse, error) {
	logger := s.logger.With(zap.String("operation", "get_by_username"), zap.String("username", in.Username))

	logger.Debug("Get user by username request")

	user, err := s.repository.GetByUsername(in.Username)
	if err != nil {
		logger.Warn("User not found", zap.Error(err))
		return nil, err
	}

	logger.Debug("User retrieved successfully", zap.Int64("user_id", int64(user.ID)))
	return toProtoUserResponse(user), nil
}

func (s *authService) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest) (*pb.AuthTokensResponse, error) {
	logger := s.logger.With(zap.String("operation", "refresh_token"))

	logger.Debug("Refresh token request")

	username, err := s.jwtParser.GetUsername(in.RefreshToken, "refresh")
	if err != nil {
		logger.Warn("Invalid refresh token", zap.Error(err))
		return nil, err
	}

	logger = logger.With(zap.String("username", username))

	user, err := s.repository.GetByUsername(username)
	if err != nil {
		logger.Warn("User not found for refresh token", zap.Error(err))
		return nil, err
	}

	userInfo := jwt.UserInfo{UserId: user.ID, Username: user.Username}
	newAccessToken, err := s.jwtGenerator.GenerateAccessToken(userInfo)
	if err != nil {
		logger.Error("Failed to generate new access token", zap.Error(err))
		return nil, err
	}

	logger.Info("Token refreshed successfully")
	return &pb.AuthTokensResponse{AccessToken: newAccessToken, RefreshToken: ""}, nil
}
