package websockets

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func (c *WebSocketConnection) ReadPump(hub *Hub, chatId string, username string, chatService services.ChatServiceInterface) {
	defer func() {
		hub.Unregister <- c
		err := c.Conn.Close()
		if err != nil {
			log.Println("Failed to read JSON:", err)
		}

	}()
	for {
		var msgContent struct {
			Content string `json:"content"`
		}
		err := c.Conn.ReadJSON(&msgContent)
		if err != nil {
			log.Println("Failed to read JSON:", err)
			break
		}

		// Nachricht mit Benutzername und Erstellungsdatum vervollständigen
		msg := models.MessageRecordDTO{
			Content:      msgContent.Content,
			Username:     username,
			CreationDate: time.Now(),
		}

		// Nachricht in der Datenbank speichern
		//if err := chatService.CreateChat(chatId, msg) {
		//	log.Println("Failed to save message to DB:", err)
		//	break
		//}

		// Nachricht an den Hub broadcasten
		hub.Broadcast <- &msg
	}
}

func (c *WebSocketConnection) WritePump() {
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
