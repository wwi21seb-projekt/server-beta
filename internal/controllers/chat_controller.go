package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
	"strconv"
)

type ChatControllerInterface interface {
	CreateChat(c *gin.Context)
	GetAllChats(c *gin.Context)
	GetChatMessages(c *gin.Context)
}

type ChatController struct {
	chatService services.ChatServiceInterface
}

func NewChatController(chatService services.ChatServiceInterface) *ChatController {
	return &ChatController{chatService: chatService}
}

func (controller *ChatController) CreateChat(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	var chatMessage models.Chat
	if err := c.ShouldBindJSON(&chatMessage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	serviceErr, httpStatus := controller.chatService.CreateChat(username.(string), chatMessage)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
}

func (controller *ChatController) GetChatMessages(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	chatMessages, serviceErr, httpStatus := controller.chatService.GetChatMessages(username.(string), offset, limit)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	response := gin.H{
		"records": chatMessages,
		"pagination": gin.H{
			"offset":  offset,
			"limit":   limit,
			"records": len(chatMessages),
		},
	}

	c.JSON(http.StatusOK, response)
}

func (controller *ChatController) GetAllChats(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	chats, serviceErr, httpStatus := controller.chatService.GetAllChats(username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(http.StatusOK, chats)
}
