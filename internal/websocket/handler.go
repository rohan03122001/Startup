// internal/websocket/handler.go

package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // In production, implement proper origin check
    },
}

type Handler struct {
    hub *Hub
    messageHandler MessageHandler
}

func NewHandler(hub *Hub) *Handler {
    return &Handler{
        hub: hub,
    }
}

func (h *Handler) SetMessageHandler(handler MessageHandler) {
    h.messageHandler = handler
}

// HandleConnection handles new WebSocket connections
func (h *Handler) HandleConnection(c *gin.Context) {
    // Upgrade HTTP connection to WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v", err)
        return
    }

    // Generate client ID
    clientID := uuid.New().String()

    // Create new client
    client := NewClient(h.hub, conn, "", clientID)

    // Set message handler
    if h.messageHandler != nil {
        client.SetMessageHandler(h.messageHandler)
    }

    // Register client with hub
    h.hub.Register <- client

    // Start client read/write routines
    go client.ReadPump()
    go client.WritePump()

    log.Printf("New WebSocket connection established: Client %s", clientID)
}