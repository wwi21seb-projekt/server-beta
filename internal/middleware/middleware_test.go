package middleware

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/utils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAuthorizeUserSuccess tests the AuthorizeUser function if it continues with the next function if the user is authorized
func TestAuthorizeUserSuccess(t *testing.T) {
	// Setup
	testUsername := "testUsername"
	validToken, err := utils.GenerateAccessToken(testUsername)
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/test", AuthorizeUser, func(c *gin.Context) {
		extractedUsername, _ := c.Get("username")
		c.String(http.StatusOK, extractedUsername.(string))
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, testUsername, w.Body.String())
}

// TestAuthorizeUserUnauthorized tests the AuthorizeUser function if it returns 404 when user does not use authentication or a valid token
func TestAuthorizeUserUnauthorized(t *testing.T) {
	invalidTokens := []string{
		"",
		"invalidToken",
	}

	for _, token := range invalidTokens {
		// Setup
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/test", AuthorizeUser, func(c *gin.Context) {
			extractedUsername, _ := c.Get("username")
			c.String(http.StatusOK, extractedUsername.(string))
		})

		// Act
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestGetLoggedInUsernameSuccess tests the GetLoggedInUsername function if it returns username and true of a valid token
func TestGetLoggedInUsernameSuccess(t *testing.T) {
	// Setup
	testUsername := "testUsername"
	validToken, err := utils.GenerateAccessToken(testUsername)
	assert.NoError(t, err)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+validToken)

	// Act and assert
	username, ok := GetLoggedInUsername(c)
	assert.True(t, ok)
	assert.Equal(t, testUsername, username)
}

// TestGetLoggedInUsernameInvalidToken tests the GetLoggedInUsername function if it returns "" and false for invalid tokens
func TestGetLoggedInUsernameInvalidToken(t *testing.T) {
	invalidTokens := []string{
		"",
		"invalidToken",
	}
	for _, token := range invalidTokens {
		// Setup
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer "+token)

		// Act and assert
		username, ok := GetLoggedInUsername(c)
		assert.False(t, ok)
		assert.Equal(t, "", username)
	}
}
