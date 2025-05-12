package handlers

import (
	"net/http"
	"strconv"

	"github.com/IAGrig/vt-csa-essays/internal/review"
	"github.com/IAGrig/vt-csa-essays/internal/review/service"
	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	service service.ReviewService
}

func NewHttpHandler(service service.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (handler *ReviewHandler) CreateReview(c *gin.Context) {
	var request review.ReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Compare the username from the auth middleware and the username specified in request
	if username, exists := c.Get("username"); !exists || username != request.Author {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can publish only your own reviews"})
		return
	}

	r, err := handler.service.Add(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, r)
}

func (handler *ReviewHandler) GetAllReviews(c *gin.Context) {
	reviews, err := handler.service.GetAllReviews()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func (handler *ReviewHandler) GetByEssayId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("essayId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reviews, err := handler.service.GetByEssayId(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func (handler *ReviewHandler) RemoveById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("reviewId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r, err := handler.service.RemoveById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, r)
}
