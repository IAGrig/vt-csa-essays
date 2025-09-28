package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

type MockReviewClient struct {
	mock.Mock
}

func (m *MockReviewClient) CreateReview(ctx context.Context, req *pb.ReviewAddRequest) (*pb.ReviewResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ReviewResponse), args.Error(1)
}

func (m *MockReviewClient) GetAllReviews(ctx context.Context, req *pb.EmptyRequest) ([]*pb.ReviewResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pb.ReviewResponse), args.Error(1)
}

func (m *MockReviewClient) GetByEssayId(ctx context.Context, req *pb.GetByEssayIdRequest) ([]*pb.ReviewResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pb.ReviewResponse), args.Error(1)
}

func (m *MockReviewClient) RemoveById(ctx context.Context, req *pb.RemoveByIdRequest) (*pb.ReviewResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ReviewResponse), args.Error(1)
}

func (m *MockReviewClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
