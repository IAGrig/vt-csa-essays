package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTParser(t *testing.T) {
	accessSecret := []byte("access-secret")
	refreshSecret := []byte("refresh-secret")
	generator := NewGenerator(accessSecret, refreshSecret)
	parser := NewParser(accessSecret, refreshSecret)

	t.Run("GetUsername_ValidTokens", func(t *testing.T) {
		tests := []struct {
			name      string
			tokenType string
			username  string
		}{
			{
				name:      "success - parses valid access token",
				tokenType: "access",
				username:  "testuser",
			},
			{
				name:      "success - parses valid refresh token",
				tokenType: "refresh",
				username:  "testuser",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var token string
				var err error

				if tt.tokenType == "access" {
					token, err = generator.GenerateAccessToken(tt.username)
				} else {
					token, err = generator.GenerateRefreshToken(tt.username)
				}
				require.NoError(t, err)

				username, err := parser.GetUsername(token, tt.tokenType)

				assert.NoError(t, err)
				assert.Equal(t, tt.username, username)
			})
		}
	})

	t.Run("GetUsername_InvalidTokens", func(t *testing.T) {
		tests := []struct {
			name        string
			token       string
			tokenType   string
			expectError bool
		}{
			{
				name:        "error - invalid token format",
				token:       "invalid.token.format",
				tokenType:   "access",
				expectError: true,
			},
			{
				name:        "error - malformed token",
				token:       "malformed.token.here",
				tokenType:   "access",
				expectError: true,
			},
			{
				name:        "error - wrong secret",
				token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0dXNlciIsInR5cGUiOiJhY2Nlc3MifQ.invalid-signature",
				tokenType:   "access",
				expectError: true,
			},
			{
				name:        "error - unknown token type",
				tokenType:   "unknown-type",
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				username, err := parser.GetUsername(tt.token, tt.tokenType)

				assert.Error(t, err)
				assert.Empty(t, username)
			})
		}
	})

	t.Run("GetUsername_WrongTokenType", func(t *testing.T) {
		accessToken, err := generator.GenerateAccessToken("testuser")
		require.NoError(t, err)

		username, err := parser.GetUsername(accessToken, "refresh")

		assert.Error(t, err)
		assert.Empty(t, username)

		refreshToken, err := generator.GenerateRefreshToken("testuser")
		require.NoError(t, err)

		username, err = parser.GetUsername(refreshToken, "access")

		assert.Error(t, err)
		assert.Empty(t, username)
	})

	t.Run("GetUsername_ExpiredToken", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":  "testuser",
			"iss":  "vt-csa-essays",
			"exp":  time.Now().Add(-1 * time.Hour).Unix(),
			"iat":  time.Now().Add(-2 * time.Hour).Unix(),
			"jti":  "test-jti",
			"type": "access",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		expiredToken, err := token.SignedString(accessSecret)
		require.NoError(t, err)

		username, err := parser.GetUsername(expiredToken, "access")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
		assert.Empty(t, username)
	})

	t.Run("GetUsername_MissingUsername", func(t *testing.T) {
		claims := jwt.MapClaims{
			"iss":  "vt-csa-essays",
			"exp":  time.Now().Add(15 * time.Minute).Unix(),
			"iat":  time.Now().Unix(),
			"jti":  "test-jti",
			"type": "access",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenStr, err := token.SignedString(accessSecret)
		require.NoError(t, err)

		username, err := parser.GetUsername(tokenStr, "access")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token doesn't contain username")
		assert.Empty(t, username)
	})

	t.Run("GetUsername_EmptyUsername", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":  "",
			"iss":  "vt-csa-essays",
			"exp":  time.Now().Add(15 * time.Minute).Unix(),
			"iat":  time.Now().Unix(),
			"jti":  "test-jti",
			"type": "access",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenStr, err := token.SignedString(accessSecret)
		require.NoError(t, err)

		username, err := parser.GetUsername(tokenStr, "access")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token doesn't contain username")
		assert.Empty(t, username)
	})

	t.Run("GetUsername_WrongTokenTypeClaim", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":  "testuser",
			"iss":  "vt-csa-essays",
			"exp":  time.Now().Add(15 * time.Minute).Unix(),
			"iat":  time.Now().Unix(),
			"jti":  "test-jti",
			"type": "wrong-type",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenStr, err := token.SignedString(accessSecret)
		require.NoError(t, err)

		username, err := parser.GetUsername(tokenStr, "access")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wrong token type")
		assert.Empty(t, username)
	})

	t.Run("GetUsername_WrongSigningMethod", func(t *testing.T) {
		wrongParser := NewParser([]byte("wrong-access-secret"), []byte("wrong-refresh-secret"))

		accessToken, err := generator.GenerateAccessToken("testuser")
		require.NoError(t, err)

		username, err := wrongParser.GetUsername(accessToken, "access")

		assert.Error(t, err)
		assert.Empty(t, username)
	})
}

func TestJWTParser_EdgeCases(t *testing.T) {
	accessSecret := []byte("access-secret")
	refreshSecret := []byte("refresh-secret")
	parser := NewParser(accessSecret, refreshSecret)

	t.Run("InvalidExpirationType", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":  "testuser",
			"iss":  "vt-csa-essays",
			"exp":  "not-a-number",
			"iat":  time.Now().Unix(),
			"jti":  "test-jti",
			"type": "access",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenStr, err := token.SignedString(accessSecret)
		require.NoError(t, err)

		username, err := parser.GetUsername(tokenStr, "access")

		assert.Error(t, err)
		assert.Empty(t, username)
	})

	t.Run("MissingTypeClaim", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub": "testuser",
			"iss": "vt-csa-essays",
			"exp": time.Now().Add(15 * time.Minute).Unix(),
			"iat": time.Now().Unix(),
			"jti": "test-jti",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenStr, err := token.SignedString(accessSecret)
		require.NoError(t, err)

		username, err := parser.GetUsername(tokenStr, "access")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wrong token type")
		assert.Empty(t, username)
	})

	t.Run("NilToken", func(t *testing.T) {
		username, err := parser.GetUsername("", "access")
		assert.Error(t, err)
		assert.Empty(t, username)
	})
}

func TestJWTIntegration(t *testing.T) {
	accessSecret := []byte("access-secret")
	refreshSecret := []byte("refresh-secret")
	generator := NewGenerator(accessSecret, refreshSecret)
	parser := NewParser(accessSecret, refreshSecret)

	t.Run("EndToEnd", func(t *testing.T) {
		username := "integrationuser"

		accessToken, err := generator.GenerateAccessToken(username)
		require.NoError(t, err)

		refreshToken, err := generator.GenerateRefreshToken(username)
		require.NoError(t, err)

		accessUsername, err := parser.GetUsername(accessToken, "access")
		assert.NoError(t, err)
		assert.Equal(t, username, accessUsername)

		refreshUsername, err := parser.GetUsername(refreshToken, "refresh")
		assert.NoError(t, err)
		assert.Equal(t, username, refreshUsername)

		// verify tokens cannot be used interchangeably
		_, err = parser.GetUsername(accessToken, "refresh")
		assert.Error(t, err)

		_, err = parser.GetUsername(refreshToken, "access")
		assert.Error(t, err)
	})

	t.Run("DifferentSecrets", func(t *testing.T) {
		generator1 := NewGenerator([]byte("secret1"), []byte("refresh1"))
		parser2 := NewParser([]byte("secret2"), []byte("refresh2"))

		token, err := generator1.GenerateAccessToken("testuser")
		require.NoError(t, err)

		username, err := parser2.GetUsername(token, "access")
		assert.Error(t, err)
		assert.Empty(t, username)
	})
}
