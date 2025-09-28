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

type TokenGenerator interface {
	GenerateAccessToken(username string) (string, error)
	GenerateRefreshToken(username string) (string, error)
}

type jwtGenerator struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewGenerator(accessSecret, refreshSecret []byte) TokenGenerator {
	return &jwtGenerator{accessSecret: accessSecret, refreshSecret: refreshSecret}
}

func (generator *jwtGenerator) GenerateAccessToken(username string) (string, error) {
	if strings.TrimSpace(username) == "" {
		return "", ErrInvalidUsername
	}

	claims := jwt.MapClaims{
		"sub":  username,
		"iss":  "vt-csa-essays",
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
		"jti":  uuid.New().String(),
		"type": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(generator.accessSecret)
}

func (generator *jwtGenerator) GenerateRefreshToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  username,
		"iss":  "vt-csa-essays",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
		"jti":  uuid.New().String(),
		"type": "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	return token.SignedString(generator.refreshSecret)
}
