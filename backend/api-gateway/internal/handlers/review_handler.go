package handlers

import (
	"net/http"
	"strconv"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/converters"
	"github.com/gin-gonic/gin"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

type ReviewHandler struct {
	reviewClient clients.ReviewClient
}

func NewReviewHandler(reviewClient clients.ReviewClient) *ReviewHandler {
	return &ReviewHandler{reviewClient: reviewClient}
}

// POST /api/reviews
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	var request struct {
		EssayId int32  `json:"essay_id" binding:"required"`
		Rank    int32  `json:"rank" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.reviewClient.CreateReview(
		c.Request.Context(),
		&pb.ReviewAddRequest{
			EssayId: request.EssayId,
			Rank:    request.Rank,
			Content: request.Content,
			Author:  username.(string),
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, converters.MarshalReviewResponse(resp))
}

// GET /api/reviews
func (h *ReviewHandler) GetAllReviews(c *gin.Context) {
	resp, err := h.reviewClient.GetAllReviews(c.Request.Context(), &pb.EmptyRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var reviews []gin.H
	for _, review := range resp {
		reviews = append(reviews, converters.MarshalReviewResponse(review))
	}

	c.JSON(http.StatusOK, reviews)
}

// GET /api/reviews/:essayId
func (h *ReviewHandler) GetByEssayId(c *gin.Context) {
	essayIdStr := c.Param("essayId")
	essayId, err := strconv.Atoi(essayIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid essay ID"})
		return
	}

	resp, err := h.reviewClient.GetByEssayId(
		c.Request.Context(),
		&pb.GetByEssayIdRequest{EssayId: int32(essayId)},
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reviews not found"})
		return
	}

	var reviews []gin.H
	for _, review := range resp {
		reviews = append(reviews, converters.MarshalReviewResponse(review))
	}

	c.JSON(http.StatusOK, reviews)
}

// DELETE /api/reviews/:reviewId
func (h *ReviewHandler) RemoveById(c *gin.Context) {
	reviewIdStr := c.Param("reviewId")
	reviewId, err := strconv.Atoi(reviewIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	resp, err := h.reviewClient.RemoveById(
		c.Request.Context(),
		&pb.RemoveByIdRequest{Id: int32(reviewId)},
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, converters.MarshalReviewResponse(resp))
}
