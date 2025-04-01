package handlers

import (
	"net/http"

	db "github.com/IAGrig/vt-csa-essays/internal/db/user"
	"github.com/IAGrig/vt-csa-essays/internal/models"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	store db.UserStore
}

func NewUserHandler(store db.UserStore) *UserHandler {
	return &UserHandler{store}
}

func (h UserHandler) CreateUser(c *gin.Context) {
	var request models.UserLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.store.Add(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h UserHandler) GetUser(c *gin.Context) {
	username := c.Param("username")

	user, err := h.store.Get(username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
