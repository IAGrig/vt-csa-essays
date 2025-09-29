package service

import (
	"context"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

type notificationService struct {
	pb.UnimplementedNotificationServiceServer
	repository repository.NotificationRepository
	logger     *logging.Logger
}

func New(repository repository.NotificationRepository, logger *logging.Logger) pb.NotificationServiceServer {
	return &notificationService{
		repository: repository,
		logger:     logger,
	}
}

func (s *notificationService) GetByUserID(in *pb.GetByUserIDRequest, stream grpc.ServerStreamingServer[pb.NotificationResponse]) error {
	logger := s.logger.With(
		zap.String("operation", "get_notifications_by_user_id"),
		zap.Int64("user_id", in.UserId),
	)

	logger.Debug("Getting notifications for user")

	notifications, err := s.repository.GetByUserID(in.UserId)
	if err != nil {
		logger.Error("Failed to get notifications from repository", zap.Error(err))
		return err
	}

	for _, notification := range notifications {
		if err := stream.Send(toProtoNotificationResponse(notification)); err != nil {
			logger.Error("Failed to send notification in stream",
				zap.Int64("notification_id", notification.NotificationID),
				zap.Error(err))
			return err
		}
	}

	logger.Debug("Notifications sent in stream", zap.Int("count", len(notifications)))
	return nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, in *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	logger := s.logger.With(
		zap.String("operation", "mark_notification_as_read"),
		zap.Int64("notification_id", in.NotificationId),
	)

	logger.Debug("Marking notification as read")

	err := s.repository.MarkAsRead(in.NotificationId)
	if err != nil {
		logger.Warn("Failed to mark notification as read", zap.Error(err))
		return &pb.MarkAsReadResponse{Success: false}, err
	}

	logger.Debug("Notification marked as read successfully")
	return &pb.MarkAsReadResponse{Success: true}, nil
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, in *pb.MarkAllAsReadRequest) (*pb.MarkAllAsReadResponse, error) {
	logger := s.logger.With(
		zap.String("operation", "mark_all_notifications_as_read"),
		zap.Int64("user_id", in.UserId),
	)

	logger.Debug("Marking all notifications as read for user")

	err := s.repository.MarkAllAsRead(in.UserId)
	if err != nil {
		logger.Error("Failed to mark all notifications as read", zap.Error(err))
		return &pb.MarkAllAsReadResponse{Success: false}, err
	}

	logger.Debug("All notifications marked as read successfully")
	return &pb.MarkAllAsReadResponse{Success: true}, nil
}
