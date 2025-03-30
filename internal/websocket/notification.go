// internal/websocket/notification.go
package websocket

import (
	"log"
	"net/http"
	"sync"

	"file-sharing-platform/internal/auth"

	"github.com/gorilla/websocket"
)

type NotificationHub struct {
	clients  map[int64][]*websocket.Conn
	mu       sync.RWMutex
	upgrader websocket.Upgrader
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients: make(map[int64][]*websocket.Conn),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in this example
				return true
			},
		},
	}
}

func (hub *NotificationHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID, err := auth.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := hub.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}

	// Register client
	hub.registerClient(userID, conn)

	// Start listening for close events
	go hub.listenForClose(userID, conn)
}

func (hub *NotificationHub) registerClient(userID int64, conn *websocket.Conn) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	hub.clients[userID] = append(hub.clients[userID], conn)
}

func (hub *NotificationHub) unregisterClient(userID int64, conn *websocket.Conn) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	// Find and remove the connection
	conns := hub.clients[userID]
	for i, c := range conns {
		if c == conn {
			// Remove this connection
			hub.clients[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	// Remove the user entry if no connections left
	if len(hub.clients[userID]) == 0 {
		delete(hub.clients, userID)
	}
}

func (hub *NotificationHub) listenForClose(userID int64, conn *websocket.Conn) {
	defer conn.Close()
	defer hub.unregisterClient(userID, conn)

	// Simple listener for close messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			// Connection closed or error
			break
		}
	}
}

func (hub *NotificationHub) NotifyUser(userID int64, message string) {
	hub.mu.RLock()
	conns := hub.clients[userID]
	hub.mu.RUnlock()

	for _, conn := range conns {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Error sending WebSocket message:", err)
			// We'll let the listen goroutine handle connection cleanup
		}
	}
}
