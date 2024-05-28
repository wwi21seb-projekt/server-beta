package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
	"strconv"
)

type MessageControllerInterface interface {
	GetMessagesByChatId(c *gin.Context)
}

type MessageController struct {
	messageService services.MessageServiceInterface
}

// NewMessageController creates a new instance of the MessageController
func NewMessageController(messageService services.MessageServiceInterface) *MessageController {
	return &MessageController{messageService: messageService}
}

// GetMessagesByChatId retrieves all messages of a chat by its chatId and can be called from the router
func (controller *MessageController) GetMessagesByChatId(c *gin.Context) {
	// Read parameters from url
	chatId := c.Param("chatId")
	offsetQuery := c.DefaultQuery("offset", "0")
	limitQuery := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		limit = 10
	}

	// Get current username
	currentUsername, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	responseDto, serviceErr, httpStatus := controller.messageService.GetMessagesByChatId(chatId, currentUsername.(string), offset, limit)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, responseDto)
}
