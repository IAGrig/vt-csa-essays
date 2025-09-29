package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/handlers"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/middleware"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"
)

func main() {
	authServicePort := os.Getenv("AUTH_SERVICE_GRPC_PORT")
	essayServicePort := os.Getenv("ESSAY_SERVICE_GRPC_PORT")
	reviewServicePort := os.Getenv("REVIEW_SERVICE_GRPC_PORT")
	notificationServicePort := os.Getenv("NOTIFICATIONS_SERVICE_GRPC_PORT")
	monitoringPort := os.Getenv("MONITORING_PORT")

	monitoring.StartMetricsServer(monitoringPort)

	authClient, err := clients.NewAuthClient("auth-service:" + authServicePort)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}
	defer authClient.Close()

	essayClient, err := clients.NewEssayClient("essay-service:" + essayServicePort)
	if err != nil {
		log.Fatalf("Failed to create essay client: %v", err)
	}
	defer essayClient.Close()

	reviewClient, err := clients.NewReviewClient("review-service:" + reviewServicePort)
	if err != nil {
		log.Fatalf("Failed to create review client: %v", err)
	}
	defer reviewClient.Close()

	notificationClient, err := clients.NewNotificationClient("notification-service:" + notificationServicePort)
	if err != nil {
		log.Fatalf("Failed to create notification client: %v", err)
	}
	defer notificationClient.Close()

	authHandler := handlers.NewAuthHandler(authClient)
	essayHandler := handlers.NewEssayHandler(essayClient)
	reviewHandler := handlers.NewReviewHandler(reviewClient)
	notificationHandler := handlers.NewNotificationHandler(notificationClient)

	router := gin.Default()

	router.Use(monitoring.GinMiddleware())

	router.Use(cors.New(cors.Config{
		AllowOrigins:	 []string{"http://localhost"},
		AllowMethods:	 []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:	 []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:	[]string{"Content-Length"},
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
		log.Fatalf("Failed to start server: %v", err)
	}
}
