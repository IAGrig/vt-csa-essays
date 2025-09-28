package handlers

import (
	"net/http"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/converters"
	"github.com/gin-gonic/gin"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
)

type EssayHandler struct {
	essayClient clients.EssayClient
}

func NewEssayHandler(essayClient clients.EssayClient) *EssayHandler {
	return &EssayHandler{essayClient: essayClient}
}

// POST /api/essays
func (h *EssayHandler) CreateEssay(c *gin.Context) {
	var request struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.essayClient.CreateEssay(
		c.Request.Context(),
		&pb.EssayAddRequest{
			Content: request.Content,
			Author: username.(string),
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, converters.MarshalProtoEssayResponse(resp))
}

// GET /api/essays/:authorname
func (h *EssayHandler) GetEssay(c *gin.Context) {
	authorname := c.Param("authorname")

	resp, err := h.essayClient.GetEssay(c.Request.Context(), &pb.GetByAuthorNameRequest{
		Authorname: authorname,
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "essay not found"})
		return
	}

	c.JSON(http.StatusOK, converters.MarshalProtoEssayWithReviewsResponse(resp))
}

// GET /api/essays
func (h *EssayHandler) GetAllEssays(c *gin.Context) {
	searchContent := c.Query("search")

	var essays []gin.H

	if searchContent == "" {
		resp, err := h.essayClient.GetAllEssays(c.Request.Context(), &pb.EmptyRequest{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, essay := range resp {
			essays = append(essays, converters.MarshalProtoEssayResponse(essay))
		}
	} else {
		resp, err := h.essayClient.SearchEssays(c.Request.Context(), &pb.SearchByContentRequest{
			Content: searchContent,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, essay := range resp {
			essays = append(essays, converters.MarshalProtoEssayResponse(essay))
		}
	}

	c.JSON(http.StatusOK, essays)
}

// DELETE /api/essays/:authorname
func (h *EssayHandler) RemoveEssay(c *gin.Context) {
	authorname := c.Param("authorname")

	usernameVal, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	usernameStr, ok := usernameVal.(string)
	if !ok || usernameStr != authorname {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can delete only your own essays"})
		return
	}


	resp, err := h.essayClient.DeleteEssay(c.Request.Context(), &pb.RemoveByAuthorNameRequest{
		Authorname: authorname,
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, converters.MarshalProtoEssayResponse(resp))
}
