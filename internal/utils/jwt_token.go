package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/models"
	"os"
	"time"
)

// GenerateJWTToken generates new jwt token with user id claim
func GenerateJWTToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	issuedAtTime := time.Now()

	claims := &jwt.MapClaims{
		"username": username,
		"exp":      expirationTime,
		"iat":      issuedAtTime, // issued at
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, err

}

// VerifyToken verifies given token and returns username
func VerifyToken(tokenString string) (*models.User, error) {
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
	initializers.DB.First(&user, claims["username"])
	if user.Username == "" {
		return nil, fmt.Errorf("user not found")
	}

	return &user, nil
}
