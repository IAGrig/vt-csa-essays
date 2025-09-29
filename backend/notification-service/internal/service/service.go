package service

import (
	"context"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/repository"
	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

type notificationService struct {
	pb.UnimplementedNotificationServiceServer
	repository repository.NotificationRepository
}

func New(repository repository.NotificationRepository) pb.NotificationServiceServer {
	return &notificationService{repository: repository}
}

func (s *notificationService) GetByUserID(in *pb.GetByUserIDRequest, stream grpc.ServerStreamingServer[pb.NotificationResponse]) error {
	notifications, err := s.repository.GetByUserID(in.UserId)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if err := stream.Send(toProtoNotificationResponse(notification)); err != nil {
			return err
		}
	}

	return nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, in *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	err := s.repository.MarkAsRead(in.NotificationId)
	if err != nil {
		return &pb.MarkAsReadResponse{Success: false}, err
	}

	return &pb.MarkAsReadResponse{Success: true}, nil
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, in *pb.MarkAllAsReadRequest) (*pb.MarkAllAsReadResponse, error) {
	err := s.repository.MarkAllAsRead(in.UserId)
	if err != nil {
		return &pb.MarkAllAsReadResponse{Success: false}, err
	}

	return &pb.MarkAllAsReadResponse{Success: true}, nil
}
