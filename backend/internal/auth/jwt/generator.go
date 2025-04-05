package jwt

import (
	"time"

	"github.com/IAGrig/vt-csa-essays/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenGenerator interface {
	GenerateAccessToken(user user.UserResponse) (string, error)
	GenerateRefreshToken(user user.UserResponse) (string, error)
}

type jwtGenerator struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewGenerator(accessSecret, refreshSecret []byte) TokenGenerator {
	return &jwtGenerator{accessSecret: accessSecret, refreshSecret: refreshSecret}
}

func (generator *jwtGenerator) GenerateAccessToken(user user.UserResponse) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.Username,
		"iss":  "vt-csa-essays",
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
		"jti":  uuid.New().String(),
		"type": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(generator.accessSecret)
}

func (generator *jwtGenerator) GenerateRefreshToken(user user.UserResponse) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.Username,
		"iss":  "vt-csa-essays",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
		"jti":  uuid.New().String(),
		"type": "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	return token.SignedString(generator.refreshSecret)
}
