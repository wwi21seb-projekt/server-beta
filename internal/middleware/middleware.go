package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/utils"
	"net/http"
	"strings"
)

// AuthorizeUser validates token and attaches username of user to request
func AuthorizeUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.PreliminaryUserUnauthorized,
		})
		return
	}

	const bearerSchema = "Bearer "
	if !strings.HasPrefix(authHeader, bearerSchema) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.PreliminaryUserUnauthorized,
		})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, bearerSchema)
	username, err := utils.VerifyAccessToken(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.PreliminaryUserUnauthorized,
		})
		return
	}

	c.Set("username", username) // Attach username to request
	c.Next()                    // Execute main function
}
