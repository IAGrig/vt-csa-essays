package handlers

import (
	"net/http"
	"os"
	"time"

	db "github.com/IAGrig/vt-csa-essays/internal/db/user"
	"github.com/IAGrig/vt-csa-essays/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	store db.UserStore
}

func NewUserHandler(store db.UserStore) *UserHandler {
	return &UserHandler{store}
}

func (h UserHandler) CreateUser(c *gin.Context) {
	var request models.UserLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.store.Add(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h UserHandler) GetUser(c *gin.Context) {
	username := c.Param("username")

	user, err := h.store.Get(username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Handles user authentication, generates access and refresh JWT tokens
// Sets the refresh token as a cookie and returns the access token in the JSON response
func (h UserHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.store.Auth(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	accessToken, err := generateAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	refreshToken, err := generateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	setRefreshCookie(c, refreshToken)
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func generateAccessToken(user models.UserResponse) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.Username,
		"iss": "vt-csa-essays",
		"exp": time.Now().Add(1 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
		"jti": uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(os.Getenv("JWT_ACCESS_SECRET")))
}

func generateRefreshToken(user models.UserResponse) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.Username,
		"iss":  "vt-csa-essays",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
		"jti":  uuid.New().String(),
		"type": "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	return token.SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))
}

func setRefreshCookie(c *gin.Context, refreshToken string) {
	isSecure := (os.Getenv("JWT_COOKIE_IS_SECURE") == "true")
	c.SetCookie(
		"refresh_token",
		refreshToken,
		7*24*60*60,      // MaxAge in seconds (7 days)
		"/auth/refresh", // Path
		"",              // Current domain
		isSecure,
		true, // HTTP only, JS can't read token
	)
}
