package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetAllChatsUnauthorized tests the GetAllChats function if it returns 401 Unauthorized when the user is not authenticated
func TestGetAllChatsUnauthorized(t *testing.T) {
	// Setup
	ctrl := NewChatController(nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/chats", nil)

	// Execution
	ctrl.GetAllChats(c)

	// Validation
	assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
}

// TestGetAllChatsSuccess tests the GetAllChats function if it returns 200 OK when chats are found
func TestGetAllChatsSuccess(t *testing.T) {
	// Setup
	ctrl := NewChatController(nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/chats", nil)
	c.Set("username", "user")

	// Execution
	ctrl.GetAllChats(c)

	// Validation
	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

// TestGetChatMessagesUnauthorized tests the GetChatMessages function if it returns 401 Unauthorized when the user is not authenticated
func TestGetChatMessagesUnauthorized(t *testing.T) {
	// Setup
	ctrl := NewChatController(nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/chats/messages", nil)

	// Execution
	ctrl.GetChatMessages(c)

	// Validation
	assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
}

// TestGetChatMessagesSuccess tests the GetChatMessages function if it returns 200 OK when chats are found
func TestGetChatMessagesSuccess(t *testing.T) {
	// Setup
	ctrl := NewChatController(nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/chats/messages", nil)
	c.Set("username", "user")

	// Execution
	ctrl.GetChatMessages(c)

	// Validation
	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

// TestGetChatMessagesNotFound tests the GetChatMessages function if it returns 404 Not Found when no chats are found
func TestGetChatMessagesNotFound(t *testing.T) {
	// Setup
	ctrl := NewChatController(nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/chats/messages", nil)
	c.Set("username", "user")

	// Execution
	ctrl.GetChatMessages(c)

	// Validation
	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
}
