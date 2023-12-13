package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/models"
	"os"
	"time"
)

// generateJWTToken generates new jwt token with user id claim
func generateJWTToken(username string, expirationTime time.Time) (string, error) {
	issuedAtTime := time.Now()

	claims := &jwt.MapClaims{
		"username": username,
		"exp":      expirationTime.Unix(),
		"iat":      issuedAtTime.Unix(), // issued at
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, err

}

// GenerateAccessToken generates new access jwt token with user id claim for 3 hours validity
func GenerateAccessToken(username string) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 3)
	tokenString, err := generateJWTToken(username, expirationTime)
	return tokenString, err

}

// VerifyAccessToken verifies given token and returns username
func VerifyAccessToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || token == nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || float64(time.Now().Unix()) > claims["exp"].(float64) {
		return nil, fmt.Errorf("expired or invalid token")
	}

	var user models.User
	initializers.DB.Where("username = ? ", claims["username"]).First(&user)
	if user.Username == "" {
		return nil, fmt.Errorf("user not found")
	}

	return &user, nil
}

// GenerateRefreshToken generates new refresh jwt token with user id claim for one week validity
func GenerateRefreshToken(username string) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 7 * 24)
	tokenString, err := generateJWTToken(username, expirationTime)
	return tokenString, err
}