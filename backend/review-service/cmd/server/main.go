package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/kafka"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/service"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

func main() {
	port := os.Getenv("REVIEW_SERVICE_GRPC_PORT")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")

	repo, err := repository.NewReviewPgRepository()
	if err != nil {
		panic(fmt.Errorf("failed to create review repository: %w", err))
	}

	brokers := strings.Split(kafkaBrokers, ",")
    producer := kafka.NewProducer(brokers, "notifications")

	reviewService := service.New(repo, producer)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterReviewServiceServer(grpcServer, reviewService)

	lis, err := net.Listen("tcp", "0.0.0.0:" + port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer.Serve(lis)
}
