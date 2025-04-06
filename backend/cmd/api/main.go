package main

import (
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

	userStore := userstore.NewUserMemStore()
	userService := userservice.New(userStore, jwtGenerator, jwtParser)
	userHandler := userhandlers.NewHttpHandler(userService)

	reviewStore := reviewstore.NewReviewMemStore()
	reviewService := reviewservice.New(reviewStore)
	reviewHandler := reviewhandlers.NewHttpHandler(reviewService)

	essayStore := essaystore.NewEssayMemStore()
	essayService := essayservice.New(essayStore, reviewStore)
	essayHandler := essayhandlers.NewHttpHandler(essayService)

	router := gin.Default()
	{
		router.POST("user", userHandler.CreateUser)
		router.POST("login", userHandler.Login)
		router.POST("refresh", userHandler.RefreshToken)

		router.GET("api/review/:essayId", reviewHandler.GetByEssayId)
	}

	publicEssayGroup := router.Group("essay")
	{
		publicEssayGroup.GET("/:authorname", essayHandler.GetEssay)
	}

	protectedEssayGroup := router.Group("essay", middleware.JWTAuthMiddleware())
	{
		protectedEssayGroup.POST("", essayHandler.CreateEssay)
		protectedEssayGroup.DELETE("/:authorname", essayHandler.RemoveEssay)
	}

	authGroup := router.Group("api", middleware.JWTAuthMiddleware())
	{
		authGroup.GET("user/:username", userHandler.GetUser)

		authGroup.POST("/review", reviewHandler.CreateReview)
		authGroup.DELETE("/review/:reviewId", reviewHandler.RemoveById)
	}

	router.Run(":8080")
}
