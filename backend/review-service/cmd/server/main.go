package main

import (
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/kafka"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/service"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"
	"go.uber.org/zap"

	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

func main() {
	logger := logging.New("review-service")
	defer logger.Sync()

	port := os.Getenv("REVIEW_SERVICE_GRPC_PORT")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	monitoringPort := os.Getenv("MONITORING_PORT")

	logger.Info("Starting review service",
		zap.String("port", port),
		zap.String("kafka_brokers", kafkaBrokers),
		zap.String("monitoring_port", monitoringPort))

	monitoring.StartMetricsServer(monitoringPort)

	repo, err := repository.NewReviewPgRepository(logger)
	if err != nil {
		logger.Fatal("Failed to create review repository",
			zap.Error(err))
	}

	brokers := strings.Split(kafkaBrokers, ",")
	producer := kafka.NewProducer(brokers, "notifications", logger)

	reviewService := service.New(repo, producer, logger)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterReviewServiceServer(grpcServer, reviewService)

	lis, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		logger.Fatal("Failed to listen",
			zap.Error(err),
			zap.String("port", port))
	}

	logger.Info("Review service starting",
		zap.String("port", port))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down review service...")
	grpcServer.GracefulStop()
	logger.Info("Review service stopped")
}
