package handlers

import (
	"net/http"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/converters"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
)

type EssayHandler struct {
	essayClient clients.EssayClient
	logger      *logging.Logger
}

func NewEssayHandler(essayClient clients.EssayClient, logger *logging.Logger) *EssayHandler {
	return &EssayHandler{
		essayClient: essayClient,
		logger:      logger,
	}
}

// POST /api/essays
func (h *EssayHandler) CreateEssay(c *gin.Context) {
	logger := h.logger.With(zap.String("operation", "create_essay"))

	var request struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warn("Invalid create essay request",
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, ok := c.Get("username")
	if !ok {
		logger.Warn("Authentication required for essay creation")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	usernameStr := username.(string)
	logger = logger.With(zap.String("username", usernameStr))

	logger.Info("Creating essay")
	resp, err := h.essayClient.CreateEssay(
		c.Request.Context(),
		&pb.EssayAddRequest{
			Content: request.Content,
			Author:  usernameStr,
		},
	)
	if err != nil {
		logger.Error("Failed to create essay",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Essay created successfully",
		zap.Int64("essay_id", int64(resp.Id)))
	c.JSON(http.StatusCreated, converters.MarshalProtoEssayResponse(resp))
}

// GET /api/essays/:authorname
func (h *EssayHandler) GetEssay(c *gin.Context) {
	authorname := c.Param("authorname")
	logger := h.logger.With(
		zap.String("operation", "get_essay"),
		zap.String("authorname", authorname),
	)

	logger.Debug("Get essay request")
	resp, err := h.essayClient.GetEssay(c.Request.Context(), &pb.GetByAuthorNameRequest{
		Authorname: authorname,
	})

	if err != nil {
		logger.Warn("Essay not found",
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "essay not found"})
		return
	}

	logger.Debug("Essay retrieved successfully")
	c.JSON(http.StatusOK, converters.MarshalProtoEssayWithReviewsResponse(resp))
}

// GET /api/essays
func (h *EssayHandler) GetAllEssays(c *gin.Context) {
	searchContent := c.Query("search")
	logger := h.logger.With(
		zap.String("operation", "get_all_essays"),
		zap.String("search_content", searchContent),
	)

	logger.Debug("Get all essays request")
	var essays []gin.H

	if searchContent == "" {
		resp, err := h.essayClient.GetAllEssays(c.Request.Context(), &pb.EmptyRequest{})
		if err != nil {
			logger.Error("Failed to get all essays",
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, essay := range resp {
			essays = append(essays, converters.MarshalProtoEssayResponse(essay))
		}
		logger.Debug("Retrieved all essays",
			zap.Int("count", len(essays)))
	} else {
		resp, err := h.essayClient.SearchEssays(c.Request.Context(), &pb.SearchByContentRequest{
			Content: searchContent,
		})
		if err != nil {
			logger.Error("Failed to search essays",
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, essay := range resp {
			essays = append(essays, converters.MarshalProtoEssayResponse(essay))
		}
		logger.Debug("Search essays completed",
			zap.Int("count", len(essays)))
	}

	c.JSON(http.StatusOK, essays)
}

// DELETE /api/essays/:authorname
func (h *EssayHandler) RemoveEssay(c *gin.Context) {
	authorname := c.Param("authorname")
	logger := h.logger.With(
		zap.String("operation", "remove_essay"),
		zap.String("authorname", authorname),
	)

	usernameVal, exists := c.Get("username")
	if !exists {
		logger.Warn("Authentication required for essay deletion")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	usernameStr, ok := usernameVal.(string)
	if !ok || usernameStr != authorname {
		logger.Warn("Forbidden essay deletion attempt",
			zap.String("authenticated_user", usernameStr))
		c.JSON(http.StatusForbidden, gin.H{"error": "you can delete only your own essays"})
		return
	}

	logger.Info("Deleting essay")
	resp, err := h.essayClient.DeleteEssay(c.Request.Context(), &pb.RemoveByAuthorNameRequest{
		Authorname: authorname,
	})

	if err != nil {
		logger.Error("Failed to delete essay",
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Essay deleted successfully")
	c.JSON(http.StatusOK, converters.MarshalProtoEssayResponse(resp))
}
