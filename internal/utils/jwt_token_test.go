package utils

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

func NewMock() {
	mockDb, _, _ := sqlmock.New()
	dialector := postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	})
	db, _ := gorm.Open(dialector, &gorm.Config{})

	initializers.DB = db
}

// TestGenerateJWTToken tests if TestGenerateJWTToken creates a token that is not empty
func TestGenerateJWTToken(t *testing.T) {
	username := "test_username"
	token, err := GenerateJWTToken(username)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)

}

func TestVerifyToken(t *testing.T) {
	// Setup
	NewMock()

	testUser := models.User{
		Username:     "test_username",
		Nickname:     "test_nickname",
		Email:        "test@test.de",
		PasswordHash: "test_hash",
		Activated:    true,
	}
	initializers.DB.Save(&testUser)
	tokenString, _ := GenerateJWTToken(testUser.Username)

	// Test valid token
	foundUser, err := VerifyAccessToken(tokenString)
	assert.Nil(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, testUser.Username, foundUser.Username)

	// Test invalid Token
	//invalidToken := "invalid_token"
	//foundUser, err := VerifyAccessToken(invalidToken)
	//assert.NotNil(t, err)
	//assert.Nil(t, foundUser.Username)
}
