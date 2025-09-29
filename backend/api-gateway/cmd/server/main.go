package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/handlers"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/middleware"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"
)

func main() {
	logger := logging.New("api-gateway")
	defer logger.Sync()

	authServicePort := os.Getenv("AUTH_SERVICE_GRPC_PORT")
	essayServicePort := os.Getenv("ESSAY_SERVICE_GRPC_PORT")
	reviewServicePort := os.Getenv("REVIEW_SERVICE_GRPC_PORT")
	notificationServicePort := os.Getenv("NOTIFICATIONS_SERVICE_GRPC_PORT")
	monitoringPort := os.Getenv("MONITORING_PORT")

	monitoring.StartMetricsServer(monitoringPort)

	authClient, err := clients.NewAuthClient("auth-service:" + authServicePort)
	if err != nil {
		logger.Fatal("Failed to create auth client", zap.Error(err))
	}
	defer authClient.Close()

	essayClient, err := clients.NewEssayClient("essay-service:" + essayServicePort)
	if err != nil {
		logger.Fatal("Failed to create essay client", zap.Error(err))
	}
	defer essayClient.Close()

	reviewClient, err := clients.NewReviewClient("review-service:" + reviewServicePort)
	if err != nil {
		logger.Fatal("Failed to create review client", zap.Error(err))
	}
	defer reviewClient.Close()

	notificationClient, err := clients.NewNotificationClient("notification-service:" + notificationServicePort)
	if err != nil {
		logger.Fatal("Failed to create notification client", zap.Error(err))
	}
	defer notificationClient.Close()

	authHandler := handlers.NewAuthHandler(authClient, logger)
	essayHandler := handlers.NewEssayHandler(essayClient, logger)
	reviewHandler := handlers.NewReviewHandler(reviewClient, logger)
	notificationHandler := handlers.NewNotificationHandler(notificationClient, logger)

	router := gin.Default()

	router.Use(monitoring.GinMiddleware())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	publicApiGroup := router.Group("/api")
	{
		authGroup := publicApiGroup.Group("/auth")
		{
			authGroup.GET("/:username", authHandler.GetUser)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/refresh", authHandler.RefreshToken)
		}

		essayGroup := publicApiGroup.Group("/essays")
		{
			essayGroup.GET("", essayHandler.GetAllEssays)
			essayGroup.GET("/:authorname", essayHandler.GetEssay)
		}

		reviewGroup := publicApiGroup.Group("/reviews")
		{
			reviewGroup.GET("", reviewHandler.GetAllReviews)
			reviewGroup.GET("/:essayId", reviewHandler.GetByEssayId)
		}
	}

	protectedApiGroup := router.Group("/api")
	protectedApiGroup.Use(middleware.JWTAuthMiddleware())
	{
		essayGroup := protectedApiGroup.Group("/essays")
		{
			essayGroup.POST("", essayHandler.CreateEssay)
			essayGroup.DELETE("/:authorname", essayHandler.RemoveEssay)
		}

		reviewGroup := protectedApiGroup.Group("/reviews")
		{
			reviewGroup.POST("", reviewHandler.CreateReview)
			reviewGroup.DELETE("/:reviewId", reviewHandler.RemoveById)
		}

		notificationGroup := protectedApiGroup.Group("/notifications")
		{
			notificationGroup.GET("", notificationHandler.GetUserNotifications)
			notificationGroup.POST("/mark-read-all", notificationHandler.MarkAllAsRead)
			notificationGroup.POST("/:notificationId/read", notificationHandler.MarkAsRead)
		}
	}

	if err := router.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
