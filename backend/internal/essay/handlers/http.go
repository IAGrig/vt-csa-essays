package handlers

import (
	"net/http"

	"github.com/IAGrig/vt-csa-essays/internal/essay"
	"github.com/IAGrig/vt-csa-essays/internal/essay/service"
	"github.com/gin-gonic/gin"
)

type EssayHandler struct {
	service service.EssaySevice
}

func NewHttpHandler(service service.EssaySevice) *EssayHandler {
	return &EssayHandler{service}
}

func (h *EssayHandler) CreateEssay(c *gin.Context) {
	var request essay.EssayRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Compare the username from the auth middleware and the username specified in request
	if username, exists := c.Get("username"); !exists || username != request.Author {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can publish only your own essays"})
		return
	}

	essay, err := h.service.Add(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, essay)
}

func (h *EssayHandler) GetEssay(c *gin.Context) {
	authorname := c.Param("authorname")

	essay, err := h.service.GetByAuthorName(authorname)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, essay)
}

func (h *EssayHandler) RemoveEssay(c *gin.Context) {
	authorname := c.Param("authorname")

	// Compare the username from the auth middleware and the authorname specified in request
	if username, exists := c.Get("username"); !exists || username != authorname {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can delete only your own essays"})
		return
	}

	essay, err := h.service.RemoveByAuthorName(authorname)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, essay)
}
