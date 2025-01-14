// internal/handlers/websocket_handler.go

package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rohan03122001/quizzing/internal/websockets"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // In production, check origin
        return true
    },
}

type WebSocketHandler struct {
    hub          *websockets.Hub
    gameHandler  *GameHandler
}

func NewWebSocketHandler(hub *websockets.Hub, gameHandler *GameHandler) *WebSocketHandler {
    return &WebSocketHandler{
        hub:         hub,
        gameHandler: gameHandler,
    }
}

// HandleConnection handles new WebSocket connections
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
    // Upgrade initial HTTP connection to WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v", err)
        return
    }

    // Generate unique client ID
    clientID := uuid.New().String()

    // Create new client
    client := websocket.NewClient(h.hub, conn, "", clientID)

    // Set message handler
    client.SetMessageHandler(h.gameHandler.HandleMessage)

    // Register client with hub
    h.hub.register <- client

    // Start client read/write pumps
    go client.ReadPump()
    go client.WritePump()

    log.Printf("New WebSocket connection established: Client %s", clientID)
}

// RegisterRoutes registers the WebSocket routes with Gin
func (h *WebSocketHandler) RegisterRoutes(router *gin.Engine) {
    router.GET("/ws", h.HandleConnection)
}