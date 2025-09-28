package service

import (
	"context"

	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/shared/jwt"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)


type authService struct {
	pb.UnimplementedUserServiceServer
	repository        repository.UserRepository
	jwtGenerator jwt.TokenGenerator
	jwtParser    jwt.TokenParser
}

func New(
	repository repository.UserRepository,
	jwtGenerator jwt.TokenGenerator,
	jwtParser jwt.TokenParser,
) pb.UserServiceServer {
	return &authService{
		repository:        repository,
		jwtGenerator: jwtGenerator,
		jwtParser:    jwtParser,
	}
}


func (s *authService) Register(ctx context.Context, in *pb.UserRegisterRequest) (*pb.UserResponse, error) {
	req := models.UserLoginRequest{Username: in.Username, Password: in.Password}
	usr, err := s.repository.Add(req)
	if err != nil {
		return nil, err
	}

	return toProtoUserResponse(usr), nil
}

func (s *authService) Auth(ctx context.Context, in *pb.UserLoginRequest) (*pb.AuthTokensResponse, error) {
	req := models.UserLoginRequest{Username: in.Username, Password: in.Password}

	user, err := s.repository.Auth(req)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.jwtGenerator.GenerateAccessToken(user.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtGenerator.GenerateRefreshToken(user.Username)
	if err != nil {
		return nil, err
	}

	resp := &pb.AuthTokensResponse{AccessToken: accessToken, RefreshToken: refreshToken}

	return resp, nil
}

func (s *authService) GetByUsername(ctx context.Context, in *pb.GetByUsernameRequest) (*pb.UserResponse, error) {
	usr, err := s.repository.GetByUsername(in.Username)
	if err != nil {
		return nil, err
	}

	return toProtoUserResponse(usr), nil
}

func (s *authService) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest) (*pb.AuthTokensResponse, error) {
	username, err := s.jwtParser.GetUsername(in.RefreshToken, "refresh")
	if err != nil {
		return nil, err
	}

	user, err := s.repository.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	newAccessToken, err := s.jwtGenerator.GenerateAccessToken(user.Username)
	if err != nil {
		return nil, err
	}

	return &pb.AuthTokensResponse{AccessToken: newAccessToken, RefreshToken: ""}, nil
}
