// internal/handlers/websocket_handler.go

package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	ws "github.com/rohan03122001/quizzing/internal/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // TODO: In production, implement proper origin check
        return true
    },
}

type WebSocketHandler struct {
    hub         *ws.Hub
    gameHandler *GameHandler
}

func NewWebSocketHandler(hub *ws.Hub, gameHandler *GameHandler) *WebSocketHandler {
    return &WebSocketHandler{
        hub:         hub,
        gameHandler: gameHandler,
    }
}

// HandleConnection handles new WebSocket connections
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
    // Upgrade HTTP connection to WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v", err)
        return
    }

    // Generate unique client ID
    clientID := uuid.New().String()

    // Create new client (initially without room)
    client := ws.NewClient(h.hub, conn, "", clientID)

    // Set message handler
    client.SetMessageHandler(h.gameHandler.HandleMessage)

    // Register with hub
    h.hub.Register <- client

    // Start read/write pumps
    go client.ReadPump()
    go client.WritePump()

    log.Printf("New WebSocket connection established: Client %s", clientID)
}

// RegisterRoutes registers the WebSocket endpoint
func (h *WebSocketHandler) RegisterRoutes(r *gin.Engine) {
    r.GET("/ws", h.HandleConnection)
}