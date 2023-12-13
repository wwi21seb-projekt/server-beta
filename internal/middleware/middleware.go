package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/utils"
	"net/http"
	"strings"
)

// AuthorizeUser validates token and attaches username of user to request
func AuthorizeUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	const bearerSchema = "Bearer "
	if !strings.HasPrefix(authHeader, bearerSchema) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, bearerSchema)
	user, err := utils.VerifyAccessToken(tokenString)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("username", user.Username) // Attach username to request
	c.Next()                         // Execute main function
}
