package controllers_test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/controllers"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/services"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestPostSubscriptionSuccess tests if PostSubscription returns 201-Created when subscription is created successfully
func TestPostSubscriptionSuccess(t *testing.T) {
	// Arrange
	mockSubscriptionService := new(services.MockSubscriptionService)
	subscriptionController := controllers.NewSubscriptionController(mockSubscriptionService)

	follower := "testFollower"
	following := "testFollowing"

	// Mock expectations
	mockSubscriptionService.On("PostSubscription", follower, following).Return(&models.SubscriptionPostResponseDTO{
		SubscriptionId: uuid.New(),
		Follower:       follower,
		Following:      following,
	}, nil, http.StatusCreated)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions", nil)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/subscriptions", subscriptionController.PostSubscription)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusCreated, w.Code) // Expect HTTP 201 Created status
}

// TestDeleteSubscriptionSuccess tests if DeleteSubscription returns 204-No Content when subscription is deleted successfully
func TestDeleteSubscriptionSuccess(t *testing.T) {
	// Arrange
	mockSubscriptionService := new(services.MockSubscriptionService)
	subscriptionController := controllers.NewSubscriptionController(mockSubscriptionService)

	subscriptionId := uuid.New()

	// Mock expectations
	mockSubscriptionService.On("DeleteSubscription", subscriptionId).Return(&models.SubscriptionDeleteResponseDTO{
		SubscriptionId: subscriptionId,
	}, nil, http.StatusNoContent)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/subscriptions/%s", subscriptionId), nil)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/subscriptions/:subscriptionId", subscriptionController.DeleteSubscription)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusNoContent, w.Code) // Expect HTTP 204 No Content status
}
