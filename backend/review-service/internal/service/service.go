package service

import (
	"context"
	"fmt"
	"log"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/kafka"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository"
	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

type reviewService struct {
	pb.UnimplementedReviewServiceServer
	repository repository.ReviewRepository
	producer kafka.Producer
	testMode bool
}

func New(repository repository.ReviewRepository, producer kafka.Producer) pb.ReviewServiceServer {
	return &reviewService{
		repository: repository,
		producer: producer,
		testMode: false,
	}
}

func NewForTest(repository repository.ReviewRepository, producer kafka.Producer) pb.ReviewServiceServer {
	return &reviewService{
		repository: repository,
		producer: producer,
		testMode: true,
	}
}

func (s *reviewService) Add(ctx context.Context, in *pb.ReviewAddRequest) (*pb.ReviewResponse, error) {
	req := models.ReviewRequest{
		EssayId: int(in.EssayId),
		Rank:    int(in.Rank),
		Content: in.Content,
		Author:  in.Author,
	}

	review, err := s.repository.Add(req)
	if err != nil {
		return nil, err
	}

	if s.producer != nil {
		event := kafka.NotificationEvent{
			Type:     "new_review",
			UserID:   0, // placeholder
			Content:  fmt.Sprintf("Your essay has been reviewed by %s", in.Author),
			EssayID:  int64(in.EssayId),
			ReviewID: int64(review.ID),
			Author:   in.Author,
		}

		if s.testMode {
			// synchronous call for tests
			if err := s.producer.SendNotificationEvent(ctx, event); err != nil {
				log.Printf("Failed to send notification event: %v", err)
			}
		} else {
			// asynchronous call for production
			go func() {
				if err := s.producer.SendNotificationEvent(context.Background(), event); err != nil {
					log.Printf("Failed to send notification event: %v", err)
				}
			}()
		}
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
