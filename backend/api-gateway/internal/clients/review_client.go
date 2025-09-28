package clients

import (
	"context"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

type ReviewClient interface {
	CreateReview(context.Context, *pb.ReviewAddRequest) (*pb.ReviewResponse, error)
	GetAllReviews(context.Context, *pb.EmptyRequest) ([]*pb.ReviewResponse, error)
	GetByEssayId(context.Context, *pb.GetByEssayIdRequest) ([]*pb.ReviewResponse, error)
	RemoveById(context.Context, *pb.RemoveByIdRequest) (*pb.ReviewResponse, error)
	Close() error
}

type reviewClient struct {
	conn    *grpc.ClientConn
	service pb.ReviewServiceClient
}

func NewReviewClient(addr string) (ReviewClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &reviewClient{
		conn:    conn,
		service: pb.NewReviewServiceClient(conn),
	}, nil
}

func (c *reviewClient) CreateReview(ctx context.Context, req *pb.ReviewAddRequest) (*pb.ReviewResponse, error) {
	return c.service.Add(ctx, req)
}

func (c *reviewClient) GetAllReviews(ctx context.Context, req *pb.EmptyRequest) ([]*pb.ReviewResponse, error) {
	stream, err := c.service.GetAllReviews(ctx, req)
	if err != nil {
		return nil, err
	}

	var reviews []*pb.ReviewResponse
	for {
		review, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (c *reviewClient) GetByEssayId(ctx context.Context, req *pb.GetByEssayIdRequest) ([]*pb.ReviewResponse, error) {
	stream, err := c.service.GetByEssayId(ctx, req)
	if err != nil {
		return nil, err
	}

	var reviews []*pb.ReviewResponse
	for {
		review, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (c *reviewClient) RemoveById(ctx context.Context, req *pb.RemoveByIdRequest) (*pb.ReviewResponse, error) {
	return c.service.RemoveById(ctx, req)
}

func (c *reviewClient) Close() error {
	return c.conn.Close()
}
