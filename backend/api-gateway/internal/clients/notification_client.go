package clients

import (
	"context"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

type NotificationClient interface {
	GetByUserID(context.Context, *pb.GetByUserIDRequest) ([]*pb.NotificationResponse, error)
	MarkAsRead(context.Context, *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error)
	MarkAllAsRead(context.Context, *pb.MarkAllAsReadRequest) (*pb.MarkAllAsReadResponse, error)
	Close() error
}

type notificationClient struct {
	conn    *grpc.ClientConn
	service pb.NotificationServiceClient
}

func NewNotificationClient(addr string) (NotificationClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &notificationClient{
		conn:    conn,
		service: pb.NewNotificationServiceClient(conn),
	}, nil
}

func (c *notificationClient) GetByUserID(ctx context.Context, req *pb.GetByUserIDRequest) ([]*pb.NotificationResponse, error) {
	stream, err := c.service.GetByUserID(ctx, req)
	if err != nil {
		return nil, err
	}

	var notifications []*pb.NotificationResponse
	for {
		notification, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (c *notificationClient) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	return c.service.MarkAsRead(ctx, req)
}

func (c *notificationClient) MarkAllAsRead(ctx context.Context, req *pb.MarkAllAsReadRequest) (*pb.MarkAllAsReadResponse, error) {
	return c.service.MarkAllAsRead(ctx, req)
}

func (c *notificationClient) Close() error {
	return c.conn.Close()
}
