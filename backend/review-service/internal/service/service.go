package service

import (
	"context"
	"fmt"
	"time"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/kafka"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

type reviewService struct {
	pb.UnimplementedReviewServiceServer
	repository repository.ReviewRepository
	producer   kafka.Producer
	logger     *logging.Logger
	testMode   bool
}

func New(repository repository.ReviewRepository, producer kafka.Producer, logger *logging.Logger) pb.ReviewServiceServer {
	return &reviewService{
		repository: repository,
		producer:   producer,
		logger:     logger,
		testMode:   false,
	}
}

func NewForTest(repository repository.ReviewRepository, producer kafka.Producer, logger *logging.Logger) pb.ReviewServiceServer {
	return &reviewService{
		repository: repository,
		producer:   producer,
		logger:     logger,
		testMode:   true,
	}
}

func (s *reviewService) Add(ctx context.Context, in *pb.ReviewAddRequest) (*pb.ReviewResponse, error) {
	start := time.Now()
	logger := s.logger.With(
		zap.String("operation", "add_review"),
		zap.Int32("essay_id", in.EssayId),
		zap.String("author", in.Author),
	)

	logger.Debug("Processing review add request")

	req := models.ReviewRequest{
		EssayId: int(in.EssayId),
		Rank:    int(in.Rank),
		Content: in.Content,
		Author:  in.Author,
	}

	review, err := s.repository.Add(req)
	if err != nil {
		monitoring.GrpcRequestDuration.WithLabelValues("review", "add", "error").Observe(float64(time.Since(start).Milliseconds()))
		monitoring.GrpcRequestsTotal.WithLabelValues("review", "add", "error").Inc()
		logger.Error("Failed to add review", zap.Error(err))
		return nil, err
	}

	duration := time.Since(start).Seconds()
	monitoring.GrpcRequestDuration.WithLabelValues("review", "Add", "success").Observe(duration)
	monitoring.GrpcRequestsTotal.WithLabelValues("review", "Add", "success").Inc()
	monitoring.ReviewsCreated.Inc()

	logger.Info("Review added successfully",
		zap.Int("review_id", review.ID),
		zap.Duration("processing_time", time.Since(start)))

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
				logger.Warn("Failed to send notification event", zap.Error(err))
			}
		} else {
			// asynchronous call for production
			go func() {
				if err := s.producer.SendNotificationEvent(context.Background(), event); err != nil {
					monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "producer_error").Inc()
					logger.Warn("Failed to send notification event asynchronously", zap.Error(err))
				}
				kafkaDuration := time.Since(kafkaStart).Seconds()
				monitoring.DbQueryDuration.WithLabelValues("kafka_produce", "notifications").Observe(kafkaDuration)
			}()
		}
	}

	return toProtoReviewResponse(review), nil
}

func (s *reviewService) GetAllReviews(in *pb.EmptyRequest, stream grpc.ServerStreamingServer[pb.ReviewResponse]) error {
	logger := s.logger.With(zap.String("operation", "get_all_reviews"))

	logger.Debug("Getting all reviews")

	reviews, err := s.repository.GetAllReviews()
	if err != nil {
		logger.Error("Failed to get all reviews", zap.Error(err))
		return err
	}

	for _, review := range reviews {
		if err := stream.Send(toProtoReviewResponse(review)); err != nil {
			logger.Error("Failed to send review in stream",
				zap.Int("review_id", review.ID),
				zap.Error(err))
			return err
		}
	}

	logger.Debug("Sent all reviews in stream", zap.Int("count", len(reviews)))
	return nil
}

func (s *reviewService) GetByEssayId(in *pb.GetByEssayIdRequest, stream grpc.ServerStreamingServer[pb.ReviewResponse]) error {
	logger := s.logger.With(
		zap.String("operation", "get_reviews_by_essay_id"),
		zap.Int32("essay_id", in.EssayId),
	)

	logger.Debug("Getting reviews by essay ID")

	reviews, err := s.repository.GetByEssayId(int(in.EssayId))
	if err != nil {
		logger.Error("Failed to get reviews by essay ID", zap.Error(err))
		return err
	}

	for _, review := range reviews {
		if err := stream.Send(toProtoReviewResponse(review)); err != nil {
			logger.Error("Failed to send review in stream",
				zap.Int("review_id", review.ID),
				zap.Error(err))
			return err
		}
	}

	logger.Debug("Sent reviews for essay in stream", zap.Int("count", len(reviews)))
	return nil
}

func (s *reviewService) RemoveById(ctx context.Context, in *pb.RemoveByIdRequest) (*pb.ReviewResponse, error) {
	logger := s.logger.With(
		zap.String("operation", "remove_review_by_id"),
		zap.Int32("review_id", in.Id),
	)

	logger.Debug("Removing review by ID")

	review, err := s.repository.RemoveById(int(in.Id))
	if err != nil {
		logger.Error("Failed to remove review", zap.Error(err))
		return nil, err
	}

	logger.Info("Review removed successfully")
	return toProtoReviewResponse(review), nil
}
