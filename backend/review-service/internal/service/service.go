package service

import (
	"context"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository"
	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

type reviewService struct {
	pb.UnimplementedReviewServiceServer
	repository repository.ReviewRepository
}

func New(repository repository.ReviewRepository) pb.ReviewServiceServer {
	return &reviewService{repository: repository}
}

func (s *reviewService) Add(ctx context.Context, in *pb.ReviewAddRequest) (*pb.ReviewResponse, error) {
	req := models.ReviewRequest{EssayId: int(in.EssayId), Rank: int(in.Rank), Content: in.Content, Author: in.Author}
	review, err := s.repository.Add(req)
	if err != nil {
		return nil, err
	}

	return toProtoReviewResponse(review), nil
}

func (s *reviewService) GetAllReviews(in *pb.EmptyRequest, stream grpc.ServerStreamingServer[pb.ReviewResponse]) error {
	reviews, err := s.repository.GetAllReviews()
	if err != nil {
		return err
	}

	for _, review := range reviews {
		if err := stream.Send(toProtoReviewResponse(review)); err != nil {
			return err
		}
	}

	return nil
}

func (s *reviewService) GetByEssayId(in *pb.GetByEssayIdRequest, stream grpc.ServerStreamingServer[pb.ReviewResponse]) error {
	reviews, err := s.repository.GetByEssayId(int(in.EssayId))
	if err != nil {
		return err
	}

	for _, review := range reviews {
		if err := stream.Send(toProtoReviewResponse(review)); err != nil {
			return err
		}
	}

	return nil
}

func (s *reviewService) RemoveById(ctx context.Context, in *pb.RemoveByIdRequest) (*pb.ReviewResponse, error) {
	review, err := s.repository.RemoveById(int(in.Id))
	if err != nil {
		return nil, err
	}

	return toProtoReviewResponse(review), nil
}
