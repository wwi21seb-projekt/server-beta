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

	// Create WebSocket connection
	// Header needs to be the same as the request header
	conn, err := controller.upgrader.Upgrade(c.Writer, c.Request, http.Header{"Sec-WebSocket-Protocol": []string{c.GetHeader("Sec-WebSocket-Protocol")}})
	if err != nil {
		return // return if connection could not be established
	}
	defer closeWebsocket(conn) // close connection when function terminates

	// Using Sec-WebSocket-Protocol header for JWT authentication because browsers do not allow custom headers
	// So middleware was not called and the JWT token needs to be verified here
	jwtToken := c.GetHeader("Sec-WebSocket-Protocol") // agreed on no Bearer prefix
	currentUsername, isRefreshToken, err := utils.VerifyJWTToken(jwtToken)
	if isRefreshToken || err != nil { // if token is a refresh token or invalid, return Unauthorized error
		sendError(conn, customerrors.UserUnauthorized)
		return // return and close connection
	}

	// Check if chat exists and if user is a participant
	_, serviceErr, _ := controller.messageService.GetChatById(chatId, currentUsername)
	if serviceErr != nil { // if no participant or chat does not exist, service returns 404 and custom error
		sendError(conn, serviceErr)
		return // return and close connection
	}

	// Add connection to map
	controller.addConnection(currentUsername, chatId, conn)
	defer controller.removeConnection(currentUsername, chatId, conn) // remove connection when function terminates

	fmt.Println("New connection for", currentUsername, "in chat", chatId)

	for {
		// Read message from client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				fmt.Println("Received close message for", currentUsername, "in chat", chatId)
				return
			}
			fmt.Println("Error reading message", err, "for user", currentUsername, "in chat", chatId)
			sendError(conn, customerrors.BadRequest)
			continue // continue to listen for more messages
		}
		fmt.Println("Message received from", currentUsername, "in chat", chatId, ":", string(message), messageType)

		// Bind message to DTO
		var req models.MessageCreateRequestDTO
		if err := json.Unmarshal(message, &req); err != nil {
			sendError(conn, customerrors.BadRequest)
			continue // continue to listen for more messages
		}

		// Get users of the chat from map that are currently connected
		// This is needed to send notifications to all other participants in the service following service function
		var connectedParticipants []string
		controller.connectionsLock.RLock()
		if controller.connections[chatId] != nil {
			for username := range controller.connections[chatId] {
				connectedParticipants = append(connectedParticipants, username)
			}
		}
		controller.connectionsLock.RUnlock()

		// Call service to save received message to database
		response, customErr, _ := controller.messageService.CreateMessage(chatId, currentUsername, &req, connectedParticipants)
		if customErr != nil {
			sendError(conn, customErr)
			continue // continue to listen for more messages
		}

		// Send message to all open connections of the chat
		responseBytes, _ := json.Marshal(response)
		controller.broadCastMessageToChat(chatId, string(responseBytes))
	}
}

// sendError sends an error message to the client using the given websocket connection
func sendError(connection *websocket.Conn, customErr *customerrors.CustomError) {
	errMessage, _ := json.Marshal(gin.H{
		"error": customErr,
	})
	err := connection.WriteMessage(websocket.TextMessage, errMessage)
	if err != nil && websocket.IsUnexpectedCloseError(err) {
		closeWebsocket(connection) // close connection if sending failed
	}
}

// closeWebsocket closes a websocket connection
func closeWebsocket(conn *websocket.Conn) {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil && websocket.IsUnexpectedCloseError(err) {
		fmt.Println("Error sending closing message:", err)
	}
	err = conn.Close()
	if err != nil && websocket.IsUnexpectedCloseError(err) {
		fmt.Println("Error closing connection:", err)
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

// removeConnection removes a connection from the map of websocket connections
func (controller *MessageController) removeConnection(username string, chatId string, conn *websocket.Conn) {
	controller.connectionsLock.Lock()
	defer controller.connectionsLock.Unlock()

	if controller.connections[chatId] == nil {
		return
	}

	connections := controller.connections[chatId][username]
	for i, c := range connections {
		if c == conn {
			fmt.Println("Removed connection for", username, "in chat", chatId)
			controller.connections[chatId][username] = append(connections[:i], connections[i+1:]...)
			break
		}
	}
	// Delete username from connections[chatId] if username has no other connections left
	if len(controller.connections[chatId][username]) == 0 {
		delete(controller.connections[chatId], username)
	}
}

// broadCastMessageToChat sends a message to all websocket connections of a chat
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
				_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				_ = c.Close() // close connection if sending failed
				controller.removeConnection(username, chatId, c)
			}
		}
	}
}
