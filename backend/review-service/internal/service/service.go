package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/kafka"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"
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
	start := time.Now()

	req := models.ReviewRequest{
		EssayId: int(in.EssayId),
		Rank:    int(in.Rank),
		Content: in.Content,
		Author:  in.Author,
	}

	review, err := s.repository.Add(req)
	if err != nil {
		monitoring.GrpcRequestDuration.WithLabelValues("review", "add", "error").Observe(float64(time.Since(start).Milliseconds()))
		monitoring.GrpcRequestsTotal.WithLabelValues("review", "add", "success").Inc()
		return nil, err
	}

	duration := time.Since(start).Seconds()
	monitoring.GrpcRequestDuration.WithLabelValues("review", "Add", "success").Observe(duration)
	monitoring.GrpcRequestsTotal.WithLabelValues("review", "Add", "success").Inc()
	monitoring.ReviewsCreated.Inc()

	if s.producer != nil {
		kafkaStart := time.Now()
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
				monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "producer_error").Inc()
				log.Printf("Failed to send notification event: %v", err)
			}
		} else {
			// asynchronous call for production
			go func() {
				if err := s.producer.SendNotificationEvent(context.Background(), event); err != nil {
					monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "producer_error").Inc()
				}
				kafkaDuration := time.Since(kafkaStart).Seconds()
				monitoring.DbQueryDuration.WithLabelValues("kafka_produce", "notifications").Observe(kafkaDuration)
log.Printf("Failed to send notification event: %v", err)
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
