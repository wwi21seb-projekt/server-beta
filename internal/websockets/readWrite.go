package websockets

import (
	"fmt"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func (c *WebSocketConnection) ReadPump(hub *Hub, chatId string, username string, messageService services.MessageServiceInterface) {
	defer func() {
		hub.Unregister <- c
		err := c.Conn.Close()
		if err != nil {
			fmt.Println("Failed to read JSON:", err)
		}

	}()
	for {
		var msgContent struct {
			Content string `json:"content"`
		}
		err := c.Conn.ReadJSON(&msgContent)
		if err != nil {
			fmt.Println("Failed to read JSON:", err)
			break
		}

		// Nachricht mit Benutzername und Erstellungsdatum vervollständigen
		msg := models.MessageRecordDTO{
			Content:      msgContent.Content,
			Username:     username,
			CreationDate: time.Now(),
		}

		// Nachricht in der Datenbank speichern
		_, err = messageService.CreateMessage(chatId, &msg)
		if err != nil {
			log.Println("Failed to save message to DB:", err)
			break
		}

		// Nachricht an den Hub broadcasten
		hub.Broadcast <- &msg
	}
}

func (c *WebSocketConnection) WritePump() {
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					fmt.Println("Failed to write message:", err)
					return
				}
				return
			}
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println("Failed to write message:", err)
				return
			}
		}
	}
}
