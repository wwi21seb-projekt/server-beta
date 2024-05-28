package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
)

type ChatControllerInterface interface {
  CreateChat(c *gin.Context)
	GetChats(c *gin.Context)
}

type ChatController struct {
	chatService services.ChatServiceInterface
}


// NewChatController can be used as a constructor to create a ChatController "object"
func NewChatController(chatService services.ChatServiceInterface) *ChatController {
	return &ChatController{chatService: chatService}
}

// CreateChat is a controller function that creates a chat for a given post id, username and the current logged-in user
func (controller *ChatController) CreateChat(c *gin.Context) {
	// Get current user from context
	currentUsername, exists := c.Get("username")
  if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}
  
  // Bind request to DTO
	var req models.ChatCreateRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
    })
    return
  }
  
  // Call service
	response, customErr, httpStatus := controller.chatService.CreatePost(&req, currentUsername.(string))
	if customErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": customErr,
		})
		return
	}

	// Return response
	c.JSON(httpStatus, response)
}

// GetChats retrieves all chats of a user by its username
func (controller *ChatController) GetChats(c *gin.Context) {
	username, exists := c.Get("username")
  if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}
  
  chats, err, status := controller.chatService.GetChatsByUsername(username.(string))
	if err != nil {
		c.JSON(status, gin.H{
			"error": err,
    })
    return
  }
  
  c.JSON(httpStatus, chats)
}
