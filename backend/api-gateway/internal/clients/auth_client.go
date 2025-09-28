package clients

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)

type AuthClient interface {
	Register(context.Context, *pb.UserRegisterRequest) (*pb.UserResponse, error)
	Login(context.Context, *pb.UserLoginRequest) (*pb.AuthTokensResponse, error)
	GetUser(context.Context, *pb.GetByUsernameRequest) (*pb.UserResponse, error)
	RefreshToken(context.Context, *pb.RefreshTokenRequest) (*pb.AuthTokensResponse, error)
	Close() error
}

type authClient struct {
	conn    *grpc.ClientConn
	service pb.UserServiceClient
}

func NewAuthClient(addr string) (AuthClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &authClient{
		conn:    conn,
		service: pb.NewUserServiceClient(conn),
	}, nil
}

func (c *authClient) Register(ctx context.Context, req *pb.UserRegisterRequest) (*pb.UserResponse, error) {
	return c.service.Register(ctx, req)
}

func (c *authClient) Login(ctx context.Context, req *pb.UserLoginRequest) (*pb.AuthTokensResponse, error) {
	return c.service.Auth(ctx, req)
}

func (c *authClient) GetUser(ctx context.Context, req *pb.GetByUsernameRequest) (*pb.UserResponse, error) {
	return c.service.GetByUsername(ctx, req)
}

func (c *authClient) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AuthTokensResponse, error) {
	return c.service.RefreshToken(ctx, req)
}

func (c *authClient) Close() error {
	return c.conn.Close()
}
