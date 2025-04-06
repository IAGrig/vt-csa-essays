package handlers

import (
	"net/http"
	"os"

	"github.com/IAGrig/vt-csa-essays/internal/user"
	"github.com/IAGrig/vt-csa-essays/internal/user/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserSevice
}

func NewHttpHandler(service service.UserSevice) *UserHandler {
	return &UserHandler{service}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var request user.UserLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Add(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	username := c.Param("username")

	user, err := h.service.GetByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Handles user authentication, generates access and refresh JWT tokens
// Sets the refresh token as a cookie and returns the access token in the JSON response
func (h *UserHandler) Login(c *gin.Context) {
	var req user.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.service.Auth(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	setRefreshCookie(c, refreshToken)
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	newAccessToken, _, err := h.service.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	// I need to implement tokens invalidation and regenerate refresh tokens on every request

	c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
}

func setRefreshCookie(c *gin.Context, refreshToken string) {
	isSecure := (os.Getenv("JWT_COOKIE_IS_SECURE") == "true")
	c.SetCookie(
		"refresh_token",
		refreshToken,
		7*24*60*60,          // MaxAge in seconds (7 days)
		"/api/user/refresh", // Path
		"",                  // Current domain
		isSecure,
		true, // HTTP only, JS can't read token
	)
}
