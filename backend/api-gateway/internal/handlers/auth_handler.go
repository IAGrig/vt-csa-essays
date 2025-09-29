package handlers

import (
	"net/http"
	"os"

	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/clients"
	"github.com/IAGrig/vt-csa-essays/backend/api-gateway/internal/converters"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)

type AuthHandler struct {
	authClient clients.AuthClient
	logger     *logging.Logger
}

func NewAuthHandler(authClient clients.AuthClient, logger *logging.Logger) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
		logger:     logger,
	}
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	logger := h.logger.With(zap.String("operation", "register"))

	var request struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warn("Invalid registration request",
			zap.Error(err),
			zap.String("username", request.Username))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("User registration attempt", zap.String("username", request.Username))

	resp, err := h.authClient.Register(
		c.Request.Context(),
		&pb.UserRegisterRequest{
			Username: request.Username,
			Password: request.Password,
		},
	)
	if err != nil {
		logger.Error("User registration failed",
			zap.String("username", request.Username),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("User registration successful",
		zap.String("username", request.Username),
		zap.Int64("user_id", int64(resp.Id)))
	c.JSON(http.StatusCreated, converters.MarshalProtoUserResponse(resp))
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	logger := h.logger.With(zap.String("operation", "login"))

	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warn("Invalid login request",
			zap.Error(err),
			zap.String("username", request.Username))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Login attempt", zap.String("username", request.Username))

	resp, err := h.authClient.Login(
		c.Request.Context(),
		&pb.UserLoginRequest{
			Username: request.Username,
			Password: request.Password,
		},
	)
	if err != nil {
		logger.Warn("Login failed - invalid credentials",
			zap.String("username", request.Username),
			zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	setRefreshCookie(c, resp.RefreshToken)
	logger.Info("Login successful", zap.String("username", request.Username))
	c.JSON(http.StatusOK, gin.H{"access_token": resp.AccessToken})
}

// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	logger := h.logger.With(zap.String("operation", "refresh_token"))

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		logger.Warn("Refresh token missing from cookie")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	logger.Debug("Refresh token attempt")

	resp, err := h.authClient.RefreshToken(
		c.Request.Context(),
		&pb.RefreshTokenRequest{RefreshToken: refreshToken},
	)
	if err != nil {
		logger.Warn("Refresh token failed",
			zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set new refresh token cookie
	setRefreshCookie(c, resp.RefreshToken)
	logger.Info("Refresh token successful")
	c.JSON(http.StatusOK, gin.H{"access_token": resp.AccessToken})
}

// GET /api/user/:username
func (h *AuthHandler) GetUser(c *gin.Context) {
	username := c.Param("username")
	logger := h.logger.With(
		zap.String("operation", "get_user"),
		zap.String("username", username),
	)

	logger.Debug("Get user request")

	resp, err := h.authClient.GetUser(
		c.Request.Context(),
		&pb.GetByUsernameRequest{Username: username},
	)
	if err != nil {
		logger.Warn("User not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	logger.Debug("User retrieved successfully")
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
