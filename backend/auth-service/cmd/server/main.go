package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/service"
	"github.com/IAGrig/vt-csa-essays/backend/shared/jwt"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)

func main() {
	port := os.Getenv("AUTH_SERVICE_GRPC_PORT")
	accessSecret := []byte(os.Getenv("JWT_ACCESS_SECRET"))
	refreshSecret := []byte(os.Getenv("JWT_REFRESH_SECRET"))
	monitoringPort := os.Getenv("MONITORING_PORT")

	monitoring.StartMetricsServer(monitoringPort)

	jwtGenerator := jwt.NewGenerator(accessSecret, refreshSecret)
	jwtParser := jwt.NewParser(accessSecret, refreshSecret)

	repo, err := repository.NewUserPgRepository()
	if err != nil {
		panic(fmt.Errorf("failed to create user repository: %w", err))
	}

	userService := service.New(repo, jwtGenerator, jwtParser)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterUserServiceServer(grpcServer, userService)

	lis, err := net.Listen("tcp", "0.0.0.0:" + port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer.Serve(lis)
}
