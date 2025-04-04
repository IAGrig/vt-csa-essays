package main

import (
	"github.com/IAGrig/vt-csa-essays/internal/db/essay"
	"github.com/IAGrig/vt-csa-essays/internal/db/user"
	"github.com/IAGrig/vt-csa-essays/internal/handlers"
	"github.com/IAGrig/vt-csa-essays/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	userStore := userstore.NewUserMemStore()
	userHandler := handlers.NewUserHandler(userStore)

	essayStore := essaystore.NewEssayMemStore()
	essayHandler := handlers.NewEssayHandler(essayStore)

	router := gin.Default()
	{
		router.POST("user", userHandler.CreateUser)
		router.POST("login", userHandler.Login)
		router.POST("refresh", userHandler.RefreshToken)
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
	}

	router.Run(":8080")
}
