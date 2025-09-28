package handlers

import (
	"net/http"
	"os"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/converters"
	"github.com/gin-gonic/gin"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)

type AuthHandler struct {
	authClient clients.AuthClient
}

func NewAuthHandler(authClient clients.AuthClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var request struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.Register(
		c.Request.Context(),
		&pb.UserRegisterRequest{
			Username: request.Username,
			Password: request.Password,
		},
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, converters.MarshalProtoUserResponse(resp))
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.Login(
		c.Request.Context(),
		&pb.UserLoginRequest{
			Username: request.Username,
			Password: request.Password,
		},
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	setRefreshCookie(c, resp.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"access_token": resp.AccessToken})
}

// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	resp, err := h.authClient.RefreshToken(
		c.Request.Context(),
		&pb.RefreshTokenRequest{RefreshToken: refreshToken},
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set new refresh token cookie
	setRefreshCookie(c, resp.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"access_token": resp.AccessToken})
}

// GET /api/user/:username
func (h *AuthHandler) GetUser(c *gin.Context) {
	username := c.Param("username")

	resp, err := h.authClient.GetUser(
		c.Request.Context(),
		&pb.GetByUsernameRequest{Username: username},
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, converters.MarshalProtoUserResponse(resp))
}

// Your existing cookie function
func setRefreshCookie(c *gin.Context, refreshToken string) {
	isSecure := (os.Getenv("JWT_COOKIE_IS_SECURE") == "true")
	c.SetCookie(
		"refresh_token",
		refreshToken,
		7*24*60*60,        // MaxAge in seconds (7 days)
		"/api/auth/refresh", // Path
		"",                 // Current domain
		isSecure,
		true, // HTTP only, JS can't read token
	)
}
