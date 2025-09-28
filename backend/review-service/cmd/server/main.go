package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/service"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

func main() {
	port := os.Getenv("REVIEW_SERVICE_GRPC_PORT")

	repo, err := repository.NewReviewPgRepository()
	if err != nil {
		panic(fmt.Errorf("failed to create review repository: %w", err))
	}

	reviewService := service.New(repo)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterReviewServiceServer(grpcServer, reviewService)

	lis, err := net.Listen("tcp", "0.0.0.0:" + port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer.Serve(lis)
}
