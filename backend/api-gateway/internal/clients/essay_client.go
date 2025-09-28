package clients

import (
	"context"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
)

type EssayClient interface {
	CreateEssay(context.Context, *pb.EssayAddRequest) (*pb.EssayResponse, error)
	GetEssay(context.Context, *pb.GetByAuthorNameRequest) (*pb.EssayWithReviewsResponse, error)
	GetAllEssays(context.Context, *pb.EmptyRequest) ([]*pb.EssayResponse, error)
	SearchEssays(context.Context, *pb.SearchByContentRequest) ([]*pb.EssayResponse, error)
	DeleteEssay(context.Context, *pb.RemoveByAuthorNameRequest) (*pb.EssayResponse, error)
	Close() error
}

type essayClient struct {
	conn    *grpc.ClientConn
	service pb.EssayServiceClient
}


func NewEssayClient(addr string) (EssayClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &essayClient{
		conn: conn,
		service: pb.NewEssayServiceClient(conn),
	}, nil
}

func (c *essayClient) CreateEssay(ctx context.Context, req *pb.EssayAddRequest) (*pb.EssayResponse, error) {
	return c.service.Add(ctx, req)
}

func (c *essayClient) GetEssay(ctx context.Context, req *pb.GetByAuthorNameRequest) (*pb.EssayWithReviewsResponse, error) {
	return c.service.GetByAuthorName(ctx, req)
}

func (c *essayClient) GetAllEssays(ctx context.Context, req *pb.EmptyRequest) ([]*pb.EssayResponse, error) {
	stream, err := c.service.GetAllEssays(ctx, req)
	if err != nil {
		return nil, err
	}

	var essays []*pb.EssayResponse
	for {
		essay, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		essays = append(essays, essay)
	}

	return essays, nil
}

func (c *essayClient) SearchEssays(ctx context.Context, req *pb.SearchByContentRequest) ([]*pb.EssayResponse, error) {
	stream, err := c.service.SearchByContent(ctx, req)
	if err != nil {
		return nil, err
	}

	var essays []*pb.EssayResponse
	for {
		essay, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		essays = append(essays, essay)
	}

	return essays, nil
}

func (c *essayClient) DeleteEssay(ctx context.Context, req *pb.RemoveByAuthorNameRequest) (*pb.EssayResponse, error) {
	return c.service.RemoveByAuthorName(ctx, req)
}

func (c *essayClient) Close() error {
	return c.conn.Close()
}
