package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"net/http"
	"strings"
)

// AuthorizeUser validates token and attaches username of user to request, if token invalid aborts with error
func AuthorizeUser(c *gin.Context) {
	username, ok := GetLoggedInUsername(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.Unauthorized,
		})
	}

	c.Set("username", username) // Attach username to request
	c.Next()                    // Execute main function
}

// GetLoggedInUsername returns the username of the logged-in user and true if the user is logged in
func GetLoggedInUsername(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}

	const bearerSchema = "Bearer "
	if !strings.HasPrefix(authHeader, bearerSchema) {
		return "", false
	}

	tokenString := strings.TrimPrefix(authHeader, bearerSchema)
	username, isRefresh, err := utils.VerifyJWTToken(tokenString)
	if err != nil || isRefresh {
		return "", false
	}

	return username, true
}
