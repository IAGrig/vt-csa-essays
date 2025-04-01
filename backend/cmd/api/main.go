package main

import (
	db "github.com/IAGrig/vt-csa-essays/internal/db/user"
	"github.com/IAGrig/vt-csa-essays/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
    userStore := db.NewUserMemStore()
    userHandler := handlers.NewUserHandler(userStore)
    router := gin.Default()

    router.POST("user", userHandler.CreateUser)
    router.GET("user/:username", userHandler.GetUser)

    router.Run(":8080")
}
