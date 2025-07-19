// pkg/auth/auth_test.go

package auth

import (
	"testing"

	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
	cfg config.Config
}

func (suite *AuthTestSuite) SetupTest() {
	suite.cfg = config.Config{
		Auth: config.AuthConfig{
			JWTSecret:   "test-secret",
			JWTTTLHours: 1,
		},
	}
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (suite *AuthTestSuite) TestGenerateAndParseToken() {
	userID := "user123"

	// Генерируем токен
	tokenStr, err := GenerateToken(suite.cfg, userID)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), tokenStr)

	// Парсим токен
	claims, err := ParseToken(suite.cfg, tokenStr)
	require.NoError(suite.T(), err)

	// Проверяем userID
	assert.Equal(suite.T(), userID, claims.UserID)

	// Проверяем Issuer
	assert.Equal(suite.T(), "GophKeeper", claims.Issuer)
}
