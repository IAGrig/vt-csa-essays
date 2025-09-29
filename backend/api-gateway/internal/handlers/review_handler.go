package handlers

import (
	"net/http"
	"strconv"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/converters"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

type ReviewHandler struct {
	reviewClient clients.ReviewClient
	logger       *logging.Logger
}

func NewReviewHandler(reviewClient clients.ReviewClient, logger *logging.Logger) *ReviewHandler {
	return &ReviewHandler{
		reviewClient: reviewClient,
		logger:       logger,
	}
}

// POST /api/reviews
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	logger := h.logger.With(zap.String("operation", "create_review"))

	var request struct {
		EssayId int32  `json:"essay_id" binding:"required"`
		Rank    int32  `json:"rank" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warn("Invalid create review request",
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		logger.Warn("Authentication required for review creation")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	usernameStr := username.(string)
	logger = logger.With(
		zap.String("username", usernameStr),
		zap.Int32("essay_id", request.EssayId),
	)

	logger.Info("Creating review")
	resp, err := h.reviewClient.CreateReview(
		c.Request.Context(),
		&pb.ReviewAddRequest{
			EssayId: request.EssayId,
			Rank:    request.Rank,
			Content: request.Content,
			Author:  usernameStr,
		},
	)
	if err != nil {
		logger.Error("Failed to create review",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Review created successfully",
		zap.Int64("review_id", int64(resp.Id)))
	c.JSON(http.StatusCreated, converters.MarshalReviewResponse(resp))
}

// GET /api/reviews
func (h *ReviewHandler) GetAllReviews(c *gin.Context) {
	logger := h.logger.With(zap.String("operation", "get_all_reviews"))

	logger.Debug("Get all reviews request")
	resp, err := h.reviewClient.GetAllReviews(c.Request.Context(), &pb.EmptyRequest{})
	if err != nil {
		logger.Error("Failed to get all reviews",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var reviews []gin.H
	for _, review := range resp {
		reviews = append(reviews, converters.MarshalReviewResponse(review))
	}

	logger.Debug("Retrieved all reviews",
		zap.Int("count", len(reviews)))
	c.JSON(http.StatusOK, reviews)
}

// GET /api/reviews/:essayId
func (h *ReviewHandler) GetByEssayId(c *gin.Context) {
	essayIdStr := c.Param("essayId")
	essayId, err := strconv.Atoi(essayIdStr)
	if err != nil {
		h.logger.Warn("Invalid essay ID",
			zap.String("essay_id", essayIdStr),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid essay ID"})
		return
	}

	logger := h.logger.With(
		zap.String("operation", "get_reviews_by_essay_id"),
		zap.Int("essay_id", essayId),
	)

	logger.Debug("Get reviews by essay ID request")
	resp, err := h.reviewClient.GetByEssayId(
		c.Request.Context(),
		&pb.GetByEssayIdRequest{EssayId: int32(essayId)},
	)
	if err != nil {
		logger.Warn("Reviews not found for essay",
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "reviews not found"})
		return
	}

	var reviews []gin.H
	for _, review := range resp {
		reviews = append(reviews, converters.MarshalReviewResponse(review))
	}

	logger.Debug("Retrieved reviews for essay",
		zap.Int("count", len(reviews)))
	c.JSON(http.StatusOK, reviews)
}

// DELETE /api/reviews/:reviewId
func (h *ReviewHandler) RemoveById(c *gin.Context) {
	reviewIdStr := c.Param("reviewId")
	reviewId, err := strconv.Atoi(reviewIdStr)
	if err != nil {
		h.logger.Warn("Invalid review ID",
			zap.String("review_id", reviewIdStr),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	logger := h.logger.With(
		zap.String("operation", "remove_review_by_id"),
		zap.Int("review_id", reviewId),
	)

	logger.Info("Deleting review")
	resp, err := h.reviewClient.RemoveById(
		c.Request.Context(),
		&pb.RemoveByIdRequest{Id: int32(reviewId)},
	)
	if err != nil {
		logger.Error("Failed to delete review",
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Review deleted successfully")
	c.JSON(http.StatusOK, converters.MarshalReviewResponse(resp))
}
