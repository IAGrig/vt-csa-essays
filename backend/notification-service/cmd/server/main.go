package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/kafka"
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/service"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"

	"google.golang.org/grpc"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/notification"
)

func main() {
	port := os.Getenv("NOTIFICATIONS_SERVICE_GRPC_PORT")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	monitoringPort := os.Getenv("MONITORING_PORT")

	monitoring.StartMetricsServer(monitoringPort)

	repo, err := repository.NewNotificationPgRepository()
	if err != nil {
		panic(fmt.Errorf("failed to create notification repository: %w", err))
	}

	brokers := strings.Split(kafkaBrokers, ",")
	consumer := kafka.NewConsumer(brokers, "notifications", "notification-service", repo)

	notificationService := service.New(repo)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterNotificationServiceServer(grpcServer, notificationService)

	lis, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go consumer.Start(ctx)

	log.Printf("Notification service started on port %s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down notification service...")

	cancel()

	grpcServer.GracefulStop()

	log.Println("Notification service stopped")
}
