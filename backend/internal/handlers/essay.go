package handlers

import (
	"net/http"

	db "github.com/IAGrig/vt-csa-essays/internal/db/essay"
	"github.com/IAGrig/vt-csa-essays/internal/models"
	"github.com/gin-gonic/gin"
)

type EssayHandler struct {
	store db.EssayStore
}

func NewEssayHandler(store db.EssayStore) *EssayHandler {
	return &EssayHandler{store}
}

func (h *EssayHandler) CreateEssay(c *gin.Context) {
	var request models.EssayRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Compare the username from the auth middleware and the username specified in request
	if username, exists := c.Get("username"); !exists || username != request.Author {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can publish only your own essays"})
		return
	}

	essay, err := h.store.Add(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, essay)
}

func (h *EssayHandler) GetEssay(c *gin.Context) {
	authorname := c.Param("authorname")

	essay, err := h.store.GetByAuthorName(authorname)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, essay)
}

func (h *EssayHandler) RemoveEssay(c *gin.Context) {
	authorname := c.Param("authorname")

	essay, err := h.store.RemoveByAuthorName(authorname)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, essay)
}
