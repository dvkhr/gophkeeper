// pkg/auth/auth_test.go

package auth

import (
	"testing"

	"github.com/dvkhr/gophkeeper/pkg/logger"
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
	logger.Logg = logger.NewTestLogger()

	suite.cfg = config.Config{
		Auth: config.AuthConfig{
			JWTSecret:           "test-secret",
			JWTTTLHours:         1,
			JWTTTLMinutes:       0,
			RefreshTokenTTLDays: 7,
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

func (suite *AuthTestSuite) TestHashPassword_ShouldHashAndPasswordCheck() {
	password := "securePassword123"

	hashed, err := HashPassword(password)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), hashed)

	// Проверяем, что пароль совпадает
	assert.True(suite.T(), CheckPasswordHash(password, hashed))

	// Проверяем, что неправильный пароль не совпадает
	assert.False(suite.T(), CheckPasswordHash("wrongPassword", hashed))
}

func (suite *AuthTestSuite) TestCheckPasswordHash_InvalidHash() {
	assert.False(suite.T(), CheckPasswordHash("password", "invalid-hash"))
}
