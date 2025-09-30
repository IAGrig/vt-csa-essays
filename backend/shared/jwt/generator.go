package jwt

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidUsername = errors.New("Invalid username")
)

type UserInfo struct {
	UserId   int
	Username string
}

type TokenGenerator interface {
	GenerateAccessToken(UserInfo) (string, error)
	GenerateRefreshToken(UserInfo) (string, error)
}

type jwtGenerator struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewGenerator(accessSecret, refreshSecret []byte) TokenGenerator {
	return &jwtGenerator{accessSecret: accessSecret, refreshSecret: refreshSecret}
}

func (generator *jwtGenerator) GenerateAccessToken(userInfo UserInfo) (string, error) {
	if strings.TrimSpace(userInfo.Username) == "" {
		return "", ErrInvalidUsername
	}

	claims := jwt.MapClaims{
		"sub":    userInfo.Username,
		"iss":    "vt-csa-essays",
		"exp":    time.Now().Add(15 * time.Minute).Unix(),
		"iat":    time.Now().Unix(),
		"jti":    uuid.New().String(),
		"type":   "access",
		"userId": userInfo.UserId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(generator.accessSecret)
}

func (generator *jwtGenerator) GenerateRefreshToken(userInfo UserInfo) (string, error) {
	if strings.TrimSpace(userInfo.Username) == "" {
		return "", ErrInvalidUsername
	}

	claims := jwt.MapClaims{
		"sub":    userInfo.Username,
		"iss":    "vt-csa-essays",
		"exp":    time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":    time.Now().Unix(),
		"jti":    uuid.New().String(),
		"type":   "refresh",
		"userId": userInfo.UserId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	return token.SignedString(generator.refreshSecret)
}
