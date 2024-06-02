package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/websockets"
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

	chats, err, httpStatus := controller.chatService.GetChatsByUsername(username.(string))
	if err != nil {
		c.JSON(httpStatus, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(httpStatus, chats)
}

// Websocket-Funktionen

// zum Verbindungsaufbau
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Globale Variablen zum Speichern der WebSocket-Verbindungen
type WebSocketConnection struct {
	Conn *websocket.Conn
}

func (controller *ChatController) HandleWebSocket(c *gin.Context) {
	chatId := c.Param("chatId")
	currentUsername, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	username := currentUsername.(string)
	err := controller.chatService.CheckUserInChat(username, chatId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
	}

	hub := websockets.GetHubManager().GetOrCreateHub(chatId)

	conn, erro := upgrader.Upgrade(c.Writer, c.Request, nil)
	if erro != nil {
		fmt.Println("Failed to upgrade to WebSocket:", err)
		return
	}

	connection := &websockets.WebSocketConnection{Conn: conn, Send: make(chan []byte, 256)}
	hub.Register <- connection

	defer func() {
		hub.Unregister <- connection
		conn.Close()
	}()

	go connection.WritePump()
	connection.ReadPump(hub, chatId, username, controller.chatService)
}
