package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/utils"
	"net/http"
)

// AuthorizeUser validates token and attaches username of user to request
func AuthorizeUser(c *gin.Context) {
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user, err := utils.VerifyToken(tokenString)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("username", user.Username) // Attach username to request
	c.Next()                         // Execute next function
}
