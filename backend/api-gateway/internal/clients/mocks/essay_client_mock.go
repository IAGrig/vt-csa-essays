package mocks

import (
	"context"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
	"github.com/stretchr/testify/mock"
)

type MockEssayClient struct {
	mock.Mock
}

func (m *MockEssayClient) CreateEssay(ctx context.Context, req *pb.EssayAddRequest) (*pb.EssayResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.EssayResponse), args.Error(1)
}

func (m *MockEssayClient) GetEssay(ctx context.Context, req *pb.GetByAuthorNameRequest) (*pb.EssayWithReviewsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.EssayWithReviewsResponse), args.Error(1)
}

func (m *MockEssayClient) GetAllEssays(ctx context.Context, req *pb.EmptyRequest) ([]*pb.EssayResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pb.EssayResponse), args.Error(1)
}

func (m *MockEssayClient) SearchEssays(ctx context.Context, req *pb.SearchByContentRequest) ([]*pb.EssayResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pb.EssayResponse), args.Error(1)
}

func (m *MockEssayClient) DeleteEssay(ctx context.Context, req *pb.RemoveByAuthorNameRequest) (*pb.EssayResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.EssayResponse), args.Error(1)
}

func (m *MockEssayClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
