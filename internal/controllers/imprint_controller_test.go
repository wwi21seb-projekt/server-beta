package controllers_test

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/controllers"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetImprint tests the GetImprint function
func TestGetImprint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	imprintController := controllers.NewImprintController()
	router.GET("/imprint", imprintController.GetImprint)

	// Create a response recorder
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/imprint", nil)
	router.ServeHTTP(w, req)

	// Parse the JSON response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check if the 'text' key exists
	if _, exists := response["text"]; !exists {
		t.Fatalf("Response does not contain 'text' key")
	}
}
