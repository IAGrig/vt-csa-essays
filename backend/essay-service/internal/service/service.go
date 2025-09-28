package service

import (
	"context"
	"fmt"
	"io"

	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/repository"
	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
	reviewPb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

type essayService struct {
	pb.UnimplementedEssayServiceServer
	essayRepository    repository.EssayRepository
	reviewClient  reviewPb.ReviewServiceClient
}

func New(essayRepository repository.EssayRepository, reviewClient reviewPb.ReviewServiceClient) pb.EssayServiceServer {
	return &essayService{essayRepository: essayRepository, reviewClient: reviewClient}
}

func (s *essayService) Add(ctx context.Context, in *pb.EssayAddRequest) (*pb.EssayResponse, error) {
	req := models.EssayRequest{Content: in.Content, Author: in.Author}
	essay, err := s.essayRepository.Add(req)
	if err != nil {
		return nil, err
	}

	return toProtoEssayResponse(essay), nil
}

func (s *essayService) GetAllEssays(in *pb.EmptyRequest, stream grpc.ServerStreamingServer[pb.EssayResponse]) error {
	essays, err := s.essayRepository.GetAllEssays()
	if err != nil {
		return err
	}

	for _, essay := range essays {
		if err := stream.Send(toProtoEssayResponse(essay)); err != nil {
			return err
		}
	}

	return nil
}

func (s *essayService) GetByAuthorName(ctx context.Context, in *pb.GetByAuthorNameRequest) (*pb.EssayWithReviewsResponse, error) {
	essay, err := s.essayRepository.GetByAuthorName(in.Authorname)
	if err != nil {
		return nil, err
	}

	reviewStream, err := s.reviewClient.GetByEssayId(ctx, &reviewPb.GetByEssayIdRequest{EssayId: int32(essay.ID)})
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews stream from review service: %w", err)
	}

	var reviews []*reviewPb.ReviewResponse
	for {
		review, err := reviewStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, review)
	}

	return toProtoEssayWithReviewsResponse(essay, reviews), nil
}

func (s *essayService) RemoveByAuthorName(ctx context.Context, in *pb.RemoveByAuthorNameRequest) (*pb.EssayResponse, error) {
	essay, err := s.essayRepository.RemoveByAuthorName(in.Authorname)
	if err != nil {
		return nil, err
	}

	return toProtoEssayResponse(essay), nil
}

func (s *essayService) SearchByContent(in *pb.SearchByContentRequest, stream grpc.ServerStreamingServer[pb.EssayResponse]) error {
	essays, err := s.essayRepository.SearchByContent(in.Content)
	if err != nil {
		return err
	}

	for _, essay := range essays {
		if err := stream.Send(toProtoEssayResponse(essay)); err != nil {
			return err
		}
	}

	return nil
}
