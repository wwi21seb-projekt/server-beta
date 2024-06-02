package websockets

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"sync"
)

type WebSocketConnection struct {
	Conn *websocket.Conn
	Send chan []byte
}

type Hub struct {
	Connections map[*WebSocketConnection]bool
	Broadcast   chan *models.MessageRecordDTO
	Register    chan *WebSocketConnection
	Unregister  chan *WebSocketConnection
	Quit        chan struct{}
	Count       int
	Mutex       sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan *models.MessageRecordDTO),
		Register:    make(chan *WebSocketConnection),
		Unregister:  make(chan *WebSocketConnection),
		Connections: make(map[*WebSocketConnection]bool),
		Quit:        make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.Mutex.Lock()
			h.Connections[conn] = true
			h.Count++
			h.Mutex.Unlock()
		case conn := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Connections[conn]; ok {
				delete(h.Connections, conn)
				close(conn.Send)
				h.Count--
				if h.Count == 0 {
					close(h.Quit)
				}
			}
			h.Mutex.Unlock()
		case message := <-h.Broadcast:
			messageBytes, err := json.Marshal(message)
			if err != nil {
				continue
			}
			for conn := range h.Connections {
				select {
				case conn.Send <- messageBytes:
				default:
					close(conn.Send)
					delete(h.Connections, conn)
				}
			}
		}
	}
}

var (
	hubManagerInstance *HubManager
	once               sync.Once
)

type HubManager struct {
	Hubs  map[string]*Hub
	Mutex sync.Mutex
}

func NewHubManager() *HubManager {
	return &HubManager{
		Hubs: make(map[string]*Hub),
	}
}

func GetHubManager() *HubManager {
	once.Do(func() {
		hubManagerInstance = NewHubManager()
	})
	return hubManagerInstance
}

func (m *HubManager) GetOrCreateHub(chatId string) *Hub {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	hub, exists := m.Hubs[chatId]
	if !exists {
		hub = NewHub()
		m.Hubs[chatId] = hub
		go func() {
			hub.Run()
			m.Mutex.Lock()
			delete(m.Hubs, chatId)
			m.Mutex.Unlock()
		}()
	}
	return hub
}
