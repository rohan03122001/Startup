// internal/websocket/client.go

package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
    // Time allowed to write a message to the peer
    writeWait = 10 * time.Second

    // Time allowed to read the next pong message from the peer
    pongWait = 60 * time.Second

    // Send pings to peer with this period (must be less than pongWait)
    pingPeriod = (pongWait * 9) / 10

    // Maximum message size allowed from peer
    maxMessageSize = 512
)

// Client represents a connected WebSocket client
type Client struct {
    // The websocket connection
    conn *websocket.Conn

    // The hub instance
    hub *Hub

    // Buffered channel of outbound messages
    send chan *GameEvent

    // Room this client is in
    RoomID string

    // Client's unique identifier
    ID string

    // Client's username
    Username string

    // Message handler function
    messageHandler func(*Client, []byte) error
    
    // Client status: "active" or "disconnected"
    Status string
    
    // Reconnection token for rejoining
    ReconnectToken string
}

// NewClient creates a new client instance
func NewClient(hub *Hub, conn *websocket.Conn, roomID, clientID string) *Client {
    return &Client{
        hub:    hub,
        conn:   conn,
        send:   make(chan *GameEvent, 256),
        RoomID: roomID,
        ID:     clientID,
        Status: "active", // Initialize as active
    }
}

// SetMessageHandler sets the function to handle incoming messages
func (c *Client) SetMessageHandler(handler func(*Client, []byte) error) {
    c.messageHandler = handler
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
    defer func() {
        // Mark client as disconnected instead of immediately unregistering
        c.hub.handleDisconnection(c)
        c.conn.Close()
        log.Printf("Client %s connection closed", c.ID)
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        // Read message
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("Unexpected close error: %v", err)
            }
            break
        }

        // Handle message using custom handler if set
        if c.messageHandler != nil {
            if err := c.messageHandler(c, message); err != nil {
                log.Printf("Error handling message from client %s: %v", c.ID, err)
                // Send error back to client
                c.hub.SendToClient(c, GameEvent{
                    Type:  "error",
                    Error: err.Error(),
                })
            }
        }
    }
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case event, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                // The hub closed the channel
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            // Write the event as JSON
            err := c.conn.WriteJSON(event)
            if err != nil {
                log.Printf("Error writing message to client %s: %v", c.ID, err)
                return
            }

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                log.Printf("Error sending ping to client %s: %v", c.ID, err)
                return
            }
        }
    }
}

// Close closes the client connection
func (c *Client) Close() {
    c.hub.Unregister <- c
    c.conn.Close()
}