package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"net/http"
	"strconv"
	"sync"
)

type MessageControllerInterface interface {
	GetMessagesByChatId(c *gin.Context)
	HandleWebSocket(c *gin.Context)
}

type MessageController struct {
	messageService services.MessageServiceInterface

	// Websockets:
	connections     map[string]map[string][]*websocket.Conn // chatId -> username -> []*websocket.Conn, for each user and chat, all connections
	connectionsLock sync.RWMutex
	upgrader        websocket.Upgrader
}

// NewMessageController creates a new instance of the MessageController
func NewMessageController(messageService services.MessageServiceInterface) *MessageController {
	return &MessageController{
		messageService: messageService,
		connections:    make(map[string]map[string][]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
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

// HandleWebSocket handles WebSocket connections for a given chatId and the logged-in user
func (controller *MessageController) HandleWebSocket(c *gin.Context) {
	// Read chatId from query parameter
	chatId := c.Query("chatId")

	// Using Sec-WebSocket-Protocol header for JWT authentication because browsers do not allow custom headers
	// So middleware was not called and the JWT token needs to be verified here
	jwtToken := c.GetHeader("Sec-WebSocket-Protocol")
	jwtToken = jwtToken[7:] // Remove "Bearer " prefix
	currentUsername, isRefreshToken, err := utils.VerifyJWTToken(jwtToken)
	if isRefreshToken || err != nil { // if token is a refresh token or invalid, return 401
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Check if chat exists and if user is a participant
	_, serviceErr, httpStatus := controller.messageService.GetMessagesByChatId(chatId, currentUsername, 0, 1)
	if serviceErr != nil { // if no participant or chat does not exist, service returns 404 and custom error
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	// Create WebSocket connection
	// Header needs to be the same as the request header
	conn, err := controller.upgrader.Upgrade(c.Writer, c.Request, http.Header{"Sec-WebSocket-Protocol": []string{c.GetHeader("Sec-WebSocket-Protocol")}})
	if err != nil {
		fmt.Println("Failed to upgrade to WebSocket for user,", currentUsername, ":", err)
		return
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close() // close connection when function terminates
	}(conn)

	// Add connection to map
	controller.addConnection(currentUsername, chatId, conn)
	defer controller.removeConnection(currentUsername, chatId, conn) // remove connection when function terminates

	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			sendError(conn, customerrors.BadRequest)
			break
		}

		// Bind message to DTO
		var req models.MessageCreateRequestDTO
		if err := json.Unmarshal(message, &req); err != nil {
			sendError(conn, customerrors.BadRequest)
			break
		}

		// Call service to save received message to database
		response, customErr, _ := controller.messageService.CreateMessage(chatId, currentUsername, &req)
		if customErr != nil {
			sendError(conn, customErr)
			break
		}

		// Send message to all open connections of the chat
		responseBytes, _ := json.Marshal(response)
		controller.broadCastMessageToChat(chatId, string(responseBytes))
	}

}

// sendError sends an error message to the client using the given connection
func sendError(connection *websocket.Conn, customErr *customerrors.CustomError) {
	errMessage, _ := json.Marshal(gin.H{
		"error": customErr,
	})
	err := connection.WriteMessage(websocket.TextMessage, errMessage)
	if err != nil {
		fmt.Println("Failed to send error websocket message:", err)
	}
}

// addConnection adds a connection to the map of connections
func (controller *MessageController) addConnection(username string, chatId string, conn *websocket.Conn) {
	controller.connectionsLock.Lock()
	defer controller.connectionsLock.Unlock()

	if controller.connections[chatId] == nil {
		controller.connections[chatId] = make(map[string][]*websocket.Conn)
	}
	controller.connections[chatId][username] = append(controller.connections[chatId][username], conn)
}

// removeConnection removes a connection from the map of connections
func (controller *MessageController) removeConnection(username string, chatId string, conn *websocket.Conn) {
	controller.connectionsLock.Lock()
	defer controller.connectionsLock.Unlock()

	if controller.connections[chatId] == nil {
		return
	}

	connections := controller.connections[chatId][username]
	for i, c := range connections {
		if c == conn {
			controller.connections[chatId][username] = append(connections[:i], connections[i+1:]...)
			break
		}
	}
}

// broadCastMessageToChat sends a message to all connections of a chat
func (controller *MessageController) broadCastMessageToChat(chatId, message string) {
	controller.connectionsLock.RLock()
	defer controller.connectionsLock.RUnlock()

	connections := controller.connections[chatId]

	// send message to all connections (also to the sender as a sending confirmation)
	// iterate through all users of the chat and then all their connections
	for username, conn := range connections {
		for _, c := range conn {
			err := c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				fmt.Println("Failed to send websocket message to ", username, ":", err)
			}
		}
	}
}
