package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Register(ctx context.Context, req *pb.UserRegisterRequest) (*pb.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *MockAuthClient) Login(ctx context.Context, req *pb.UserLoginRequest) (*pb.AuthTokensResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AuthTokensResponse), args.Error(1)
}

func (m *MockAuthClient) GetUser(ctx context.Context, req *pb.GetByUsernameRequest) (*pb.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AuthTokensResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AuthTokensResponse), args.Error(1)
}

func (m *MockAuthClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
