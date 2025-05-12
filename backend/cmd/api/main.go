package main

import (
	"fmt"
	"os"

	"github.com/IAGrig/vt-csa-essays/internal/auth/jwt"
	"github.com/IAGrig/vt-csa-essays/internal/auth/middleware"
	essayhandlers "github.com/IAGrig/vt-csa-essays/internal/essay/handlers"
	essayservice "github.com/IAGrig/vt-csa-essays/internal/essay/service"
	essaystore "github.com/IAGrig/vt-csa-essays/internal/essay/store"
	reviewhandlers "github.com/IAGrig/vt-csa-essays/internal/review/handlers"
	reviewservice "github.com/IAGrig/vt-csa-essays/internal/review/service"
	reviewstore "github.com/IAGrig/vt-csa-essays/internal/review/store"
	userhandlers "github.com/IAGrig/vt-csa-essays/internal/user/handlers"
	userservice "github.com/IAGrig/vt-csa-essays/internal/user/service"
	userstore "github.com/IAGrig/vt-csa-essays/internal/user/store"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	accessSecret := []byte(os.Getenv("JWT_ACCESS_SECRET"))
	refreshSecret := []byte(os.Getenv("JWT_REFRESH_SECRET"))

	jwtGenerator := jwt.NewGenerator(accessSecret, refreshSecret)
	jwtParser := jwt.NewParser(accessSecret, refreshSecret)

	userStore, err := userstore.NewUserPgStore()
	if err != nil {
		panic(fmt.Errorf("failed to create user store: %w", err))
	}

	reviewStore, err := reviewstore.NewReviewPgStore()
	if err != nil {
		panic(fmt.Errorf("failed to create review store: %w", err))
	}

	essayStore, err := essaystore.NewEssayPgStore()
	if err != nil {
		panic(fmt.Errorf("failed to create essay store: %w", err))
	}

	userService := userservice.New(userStore, jwtGenerator, jwtParser)
	userHandler := userhandlers.NewHttpHandler(userService)

	reviewService := reviewservice.New(reviewStore)
	reviewHandler := reviewhandlers.NewHttpHandler(reviewService)

	essayService := essayservice.New(essayStore, reviewService)
	essayHandler := essayhandlers.NewHttpHandler(essayService)

	router := gin.Default()
	publicApiGroup := router.Group("api")
	{
		userGroup := publicApiGroup.Group("user")
		{
			userGroup.GET("/:username", userHandler.GetUser)
			userGroup.POST("", userHandler.CreateUser)
			userGroup.POST("login", userHandler.Login)
			userGroup.POST("refresh", userHandler.RefreshToken)
		}

		essayGroup := publicApiGroup.Group("essay")
		{
			essayGroup.GET("", essayHandler.GetAllEssays)
			essayGroup.GET("/:authorname", essayHandler.GetEssay)
		}

		reviewGroup := publicApiGroup.Group("review")
		{
			reviewGroup.GET("", reviewHandler.GetAllReviews)
			reviewGroup.GET("/:essayId", reviewHandler.GetByEssayId)
		}
	}

	protectedApiGroup := router.Group("api", middleware.JWTAuthMiddleware())
	{
		essayGroup := protectedApiGroup.Group("essay")
		{
			essayGroup.POST("", essayHandler.CreateEssay)
			essayGroup.DELETE("/:authorname", essayHandler.RemoveEssay)
		}

		reviewGroup := protectedApiGroup.Group("review")
		{
			reviewGroup.POST("", reviewHandler.CreateReview)
			reviewGroup.DELETE("/:reviewId", reviewHandler.RemoveById)
		}
	}

	router.Run(":8080")
}
