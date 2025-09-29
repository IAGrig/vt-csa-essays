package service

import (
	"context"
	"fmt"
	"io"

	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
	reviewPb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

type essayService struct {
	pb.UnimplementedEssayServiceServer
	essayRepository repository.EssayRepository
	reviewClient    reviewPb.ReviewServiceClient
	logger          *logging.Logger
}

func New(essayRepository repository.EssayRepository, reviewClient reviewPb.ReviewServiceClient, logger *logging.Logger) pb.EssayServiceServer {
	return &essayService{
		essayRepository: essayRepository,
		reviewClient:    reviewClient,
		logger:          logger,
	}
}

func (s *essayService) Add(ctx context.Context, in *pb.EssayAddRequest) (*pb.EssayResponse, error) {
	logger := s.logger.With(
		zap.String("operation", "add_essay"),
		zap.String("author", in.Author),
	)

	logger.Info("Adding new essay")

	req := models.EssayRequest{Content: in.Content, Author: in.Author}
	essay, err := s.essayRepository.Add(req)
	if err != nil {
		logger.Error("Failed to add essay", zap.Error(err))
		return nil, err
	}

	logger.Info("Essay added successfully", zap.Int64("essay_id", int64(essay.ID)))
	return toProtoEssayResponse(essay), nil
}

func (s *essayService) GetAllEssays(in *pb.EmptyRequest, stream grpc.ServerStreamingServer[pb.EssayResponse]) error {
	logger := s.logger.With(zap.String("operation", "get_all_essays"))

	logger.Debug("Getting all essays")

	essays, err := s.essayRepository.GetAllEssays()
	if err != nil {
		logger.Error("Failed to get all essays", zap.Error(err))
		return err
	}

	for _, essay := range essays {
		if err := stream.Send(toProtoEssayResponse(essay)); err != nil {
			logger.Error("Failed to send essay in stream", zap.Error(err))
			return err
		}
	}

	logger.Debug("Sent all essays in stream", zap.Int("count", len(essays)))
	return nil
}

func (s *essayService) GetByAuthorName(ctx context.Context, in *pb.GetByAuthorNameRequest) (*pb.EssayWithReviewsResponse, error) {
	logger := s.logger.With(
		zap.String("operation", "get_essay_by_author"),
		zap.String("author", in.Authorname),
	)

	logger.Debug("Getting essay by author name with reviews")

	essay, err := s.essayRepository.GetByAuthorName(in.Authorname)
	if err != nil {
		logger.Warn("Essay not found", zap.Error(err))
		return nil, err
	}

	logger = logger.With(zap.Int64("essay_id", int64(essay.ID)))

	reviewStream, err := s.reviewClient.GetByEssayId(ctx, &reviewPb.GetByEssayIdRequest{EssayId: int32(essay.ID)})
	if err != nil {
		logger.Error("Failed to get reviews stream from review service", zap.Error(err))
		return nil, fmt.Errorf("failed to get reviews stream from review service: %w", err)
	}

	var reviews []*reviewPb.ReviewResponse
	for {
		review, err := reviewStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("Failed to receive review from stream", zap.Error(err))
			return nil, err
		}

		reviews = append(reviews, review)
	}

	logger.Debug("Retrieved essay with reviews", zap.Int("reviews_count", len(reviews)))
	return toProtoEssayWithReviewsResponse(essay, reviews), nil
}

func (s *essayService) RemoveByAuthorName(ctx context.Context, in *pb.RemoveByAuthorNameRequest) (*pb.EssayResponse, error) {
	logger := s.logger.With(
		zap.String("operation", "remove_essay_by_author"),
		zap.String("author", in.Authorname),
	)

	logger.Info("Removing essay by author name")

	essay, err := s.essayRepository.RemoveByAuthorName(in.Authorname)
	if err != nil {
		logger.Error("Failed to remove essay", zap.Error(err))
		return nil, err
	}

	logger.Info("Essay removed successfully", zap.Int64("essay_id", int64(essay.ID)))
	return toProtoEssayResponse(essay), nil
}

func (s *essayService) SearchByContent(in *pb.SearchByContentRequest, stream grpc.ServerStreamingServer[pb.EssayResponse]) error {
	logger := s.logger.With(
		zap.String("operation", "search_essays_by_content"),
		zap.String("search_term", in.Content),
	)

	logger.Debug("Searching essays by content")

	essays, err := s.essayRepository.SearchByContent(in.Content)
	if err != nil {
		logger.Error("Failed to search essays", zap.Error(err))
		return err
	}

	for _, essay := range essays {
		if err := stream.Send(toProtoEssayResponse(essay)); err != nil {
			logger.Error("Failed to send essay in search stream", zap.Error(err))
			return err
		}
	}

	logger.Debug("Search completed and results sent", zap.Int("results_count", len(essays)))
	return nil
}
