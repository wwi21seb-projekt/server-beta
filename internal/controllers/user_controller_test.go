package controllers

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setup() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/user", CreateUser)
	return router
}

// TestCreateUserInvalidBody tests if CreateUser returns 400-Bad Request when request body/username/... is invalid
func TestCreateUserBadRequest(t *testing.T) {

	router := setup()

	invalidBodies := []string{
		`{"invalidField": "value"}`, // invalid body
		`{"username": "", "nickname": "", "password": "Password123!", "email": "email_test@testdomain.de"}`,   // no username
		`{"username": "testUser", "nickname": "", "password": "passwd", "email": "email_test@testdomain.de"}`, // password does not meet specifications
		`{"username": "testUser", "nickname": "", "password": "passwd123!", "email": "testDomain.de"}`,        // invalid email syntax
	}

	for _, body := range invalidBodies {
		// Create request
		req, err := http.NewRequest(http.MethodPost, "/user", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		var errorResponse errors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		expectedCustomError := errors.BadRequest
		if errorResponse.Error.Message != expectedCustomError.Message && errorResponse.Error.Code != expectedCustomError.Code {
			t.Errorf("Expected error message '%s' and code '%s', got message '%s' and code '%s'", expectedCustomError.Message, expectedCustomError.Code, errorResponse.Error.Message, errorResponse.Error.Code)
		}
	}
}
