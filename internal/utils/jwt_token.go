package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

// generateJWTToken generates new jwt token with user id claim
func generateJWTToken(username string, expirationTime time.Time, isRefreshToken bool) (string, error) {
	issuedAtTime := time.Now()

	claims := &jwt.MapClaims{
		"username": username,
		"exp":      expirationTime.Unix(),
		"iat":      issuedAtTime.Unix(), // issued at
		"refresh":  isRefreshToken,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, err

}

// GenerateAccessToken generates new access jwt token with user id claim for 3 hours validity
func GenerateAccessToken(username string) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 3)
	tokenString, err := generateJWTToken(username, expirationTime, false)
	return tokenString, err
}

// GenerateRefreshToken generates new refresh jwt token with user id claim for one week validity
func GenerateRefreshToken(username string) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 7 * 24)
	tokenString, err := generateJWTToken(username, expirationTime, true)
	return tokenString, err
}

// VerifyJWTToken verifies given token and returns username and true if token is refresh token
func VerifyJWTToken(tokenString string) (string, bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || token == nil || !token.Valid {
		return "", false, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || float64(time.Now().Unix()) > claims["exp"].(float64) {
		return "", false, fmt.Errorf("expired or invalid token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", false, fmt.Errorf("invalid token")
	}
	isRefreshToken, ok := claims["refresh"].(bool)
	if !ok {
		return "", false, fmt.Errorf("invalid token")
	}

	return username, isRefreshToken, nil
}
