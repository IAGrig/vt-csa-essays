package mocks

import (
	"context"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
	"github.com/stretchr/testify/mock"
)

type MockNotificationClient struct {
	mock.Mock
}

func (m *MockNotificationClient) GetByUserID(ctx context.Context, req *pb.GetByUserIDRequest) ([]*pb.NotificationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pb.NotificationResponse), args.Error(1)
}

func (m *MockNotificationClient) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.MarkAsReadResponse), args.Error(1)
}

func (m *MockNotificationClient) MarkAllAsRead(ctx context.Context, req *pb.MarkAllAsReadRequest) (*pb.MarkAllAsReadResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.MarkAllAsReadResponse), args.Error(1)
}

func (m *MockNotificationClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
