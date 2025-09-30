package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTGenerator(t *testing.T) {
	accessSecret := []byte("access-secret")
	refreshSecret := []byte("refresh-secret")
	generator := NewGenerator(accessSecret, refreshSecret)

	t.Run("GenerateAccessToken", func(t *testing.T) {
		tests := []struct {
			name     string
			userId   int
			username string
			wantErr  bool
		}{
			{
				name:     "success - generates valid access token",
				userId:   1,
				username: "testuser",
				wantErr:  false,
			},
			{
				name:     "fail - generates token for empty username",
				userId:   1,
				username: "",
				wantErr:  true,
			},
			{
				name:     "fail - generates token for whitespace username",
				userId:   1,
				username: "    ",
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				token, err := generator.GenerateAccessToken(UserInfo{UserId: tt.userId, Username: tt.username})

				if tt.wantErr {
					assert.Error(t, err)
					assert.Empty(t, token)
				} else {
					assert.NoError(t, err)
					assert.NotEmpty(t, token)

					parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
						return accessSecret, nil
					})
					require.NoError(t, err)
					assert.True(t, parsedToken.Valid)

					claims, ok := parsedToken.Claims.(jwt.MapClaims)
					require.True(t, ok)

					assert.Equal(t, tt.username, claims["sub"])
					assert.Equal(t, "vt-csa-essays", claims["iss"])
					assert.Equal(t, "access", claims["type"])
					assert.NotEmpty(t, claims["jti"])
					assert.NotZero(t, claims["iat"])
					assert.NotZero(t, claims["exp"])
					assert.EqualValues(t, float64(tt.userId), claims["userId"])

					exp := time.Unix(int64(claims["exp"].(float64)), 0)
					expectedExp := time.Now().Add(15 * time.Minute)
					assert.WithinDuration(t, expectedExp, exp, time.Second)
				}
			})
		}
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		tests := []struct {
			name     string
			userId   int
			username string
			wantErr  bool
		}{
			{
				name:     "success - generates valid refresh token",
				userId:   1,
				username: "testuser",
				wantErr:  false,
			},
			{
				name:     "success - generates token for empty username",
				userId:   1,
				username: "",
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				token, err := generator.GenerateRefreshToken(UserInfo{UserId: tt.userId, Username: tt.username})

				if tt.wantErr {
					assert.Error(t, err)
					assert.Empty(t, token)
				} else {
					assert.NoError(t, err)
					assert.NotEmpty(t, token)

					parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
						return refreshSecret, nil
					})
					require.NoError(t, err)
					assert.True(t, parsedToken.Valid)

					claims, ok := parsedToken.Claims.(jwt.MapClaims)
					require.True(t, ok)

					assert.Equal(t, tt.username, claims["sub"])
					assert.Equal(t, "vt-csa-essays", claims["iss"])
					assert.Equal(t, "refresh", claims["type"])
					assert.NotEmpty(t, claims["jti"])
					assert.NotZero(t, claims["iat"])
					assert.NotZero(t, claims["exp"])
					assert.EqualValues(t, float64(tt.userId), claims["userId"])

					exp := time.Unix(int64(claims["exp"].(float64)), 0)
					expectedExp := time.Now().Add(7 * 24 * time.Hour)
					assert.WithinDuration(t, expectedExp, exp, time.Second)
				}
			})
		}
	})

	t.Run("DifferentSigningMethods", func(t *testing.T) {
		accessToken, err := generator.GenerateAccessToken(UserInfo{UserId: 1, Username: "testuser"})
		require.NoError(t, err)

		refreshToken, err := generator.GenerateRefreshToken(UserInfo{UserId: 1, Username: "testuser"})
		require.NoError(t, err)

		accessParsed, _ := jwt.Parse(accessToken, nil)
		refreshParsed, _ := jwt.Parse(refreshToken, nil)

		assert.Equal(t, "HS512", accessParsed.Method.Alg())
		assert.Equal(t, "HS384", refreshParsed.Method.Alg())
	})

	t.Run("TokenUniqueness", func(t *testing.T) {
		token1, err := generator.GenerateAccessToken(UserInfo{UserId: 1, Username: "testuser"})
		require.NoError(t, err)

		token2, err := generator.GenerateAccessToken(UserInfo{UserId: 1, Username: "testuser"})
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2, "Tokens should be different due to unique jti")
	})
}
