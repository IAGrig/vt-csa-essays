package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/service"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
	reviewPb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

func main() {
	port := os.Getenv("ESSAY_SERVICE_GRPC_PORT")
	reviewServicePort := os.Getenv("REVIEW_SERVICE_GRPC_PORT")
	monitoringPort := os.Getenv("MONITORING_PORT")

	monitoring.StartMetricsServer(monitoringPort)

	repo, err := repository.NewEssayPgRepository()
	if err != nil {
		panic(fmt.Errorf("failed to create essay repository: %w", err))
	}

	reviewConn, err := grpc.NewClient(
		"review-service:" + reviewServicePort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to review service: %v", err)
	}
	defer reviewConn.Close()

	reviewClient := reviewPb.NewReviewServiceClient(reviewConn)

	essayService := service.New(repo, reviewClient)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEssayServiceServer(grpcServer, essayService)

	lis, err := net.Listen("tcp", "0.0.0.0:" + port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer.Serve(lis)
}
