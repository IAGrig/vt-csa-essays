package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenParser interface {
	GetUsername(token, tokenType string) (string, error)
	GetUserId(token, tokenType string) (int, error)
}

type jwtParser struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewParser(accessSecret, refreshSecret []byte) TokenParser {
	return &jwtParser{accessSecret: accessSecret, refreshSecret: refreshSecret}
}

func (parser *jwtParser) GetUsername(tokenStr, tokenType string) (string, error) {
	var secret []byte
	switch tokenType {
	case "access":
		secret = parser.accessSecret
	case "refresh":
		secret = parser.refreshSecret
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != tokenType {
		return "", fmt.Errorf("wrong token type")
	}

	// Validate token expiration
	exp, ok := claims["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		return "", fmt.Errorf("token is expired")
	}

	username, ok := claims["sub"].(string)
	if !ok || username == "" {
		return "", fmt.Errorf("token doesn't contain username")
	}

	return username, nil
}

func (parser *jwtParser) GetUserId(tokenStr, tokenType string) (int, error) {
	var secret []byte
	switch tokenType {
	case "access":
		secret = parser.accessSecret
	case "refresh":
		secret = parser.refreshSecret
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != tokenType {
		return 0, fmt.Errorf("wrong token type")
	}

	// Validate token expiration
	exp, ok := claims["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		return 0, fmt.Errorf("token is expired")
	}

	userId, ok := claims["userId"].(float64)
	if !ok || userId == 0 {
		return 0, fmt.Errorf("token doesn't contain userId")
	}

	return int(userId), nil
}
