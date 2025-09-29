package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/kafka"
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/service"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"
	"go.uber.org/zap"

	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

func main() {
	logger := logging.New("notification-service")
	defer logger.Sync()

	port := os.Getenv("NOTIFICATIONS_SERVICE_GRPC_PORT")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	monitoringPort := os.Getenv("MONITORING_PORT")

	logger.Info("Starting notification service",
		zap.String("port", port),
		zap.String("kafka_brokers", kafkaBrokers),
		zap.String("monitoring_port", monitoringPort))

	monitoring.StartMetricsServer(monitoringPort)

	repo, err := repository.NewNotificationPgRepository(logger)
	if err != nil {
		logger.Fatal("Failed to create notification repository",
			zap.Error(err))
	}

	brokers := strings.Split(kafkaBrokers, ",")
	consumer := kafka.NewConsumer(brokers, "notifications", "notification-service", repo, logger)

	notificationService := service.New(repo, logger)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterNotificationServiceServer(grpcServer, notificationService)

	lis, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		logger.Fatal("Failed to listen",
			zap.Error(err),
			zap.String("port", port))
	}

	logger.Info("Notification service starting",
			zap.String("port", port))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go consumer.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down notification service...")
	cancel()
	grpcServer.GracefulStop()
	logger.Info("Notification service stopped")
}
