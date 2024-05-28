package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
)

type ChatControllerInterface interface {
	GetChats(c *gin.Context)
}

type ChatController struct {
	chatService services.ChatServiceInterface
}

func NewChatController(chatService services.ChatServiceInterface) *ChatController {
	return &ChatController{chatService: chatService}
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

	c.JSON(http.StatusOK, chats)
}
