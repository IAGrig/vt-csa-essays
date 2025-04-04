package main

import (
	db "github.com/IAGrig/vt-csa-essays/internal/db/user"
	"github.com/IAGrig/vt-csa-essays/internal/handlers"
	"github.com/IAGrig/vt-csa-essays/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	userStore := db.NewUserMemStore()
	userHandler := handlers.NewUserHandler(userStore)

	router := gin.Default()
	{
		router.POST("user", userHandler.CreateUser)
		router.POST("login", userHandler.Login)
		router.POST("refresh", userHandler.RefreshToken)
	}

	authGroup := router.Group("api", middleware.JWTAuthMiddleware())
	{
		authGroup.GET("user/:username", userHandler.GetUser)
	}

	router.Run(":8080")
}
