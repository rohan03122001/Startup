// internal/websocket/client.go

package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
    writeWait = 10 * time.Second
    pongWait = 60 * time.Second
    pingPeriod = (pongWait * 9) / 10
    maxMessageSize = 512
)

// MessageHandler is a function type for handling incoming messages
type MessageHandler func(*Client, []byte) error

// Client represents a connected WebSocket client
type Client struct {
    hub *Hub

    // The websocket connection
    conn *websocket.Conn

    // Buffered channel of outbound messages
    send chan *GameEvent

    // Room ID this client is in
    RoomID string

    // Unique identifier for this client
    ID string

    // Client's username
    Username string

    // Message handler function
    messageHandler MessageHandler
}

func NewClient(hub *Hub, conn *websocket.Conn, roomID string, id string) *Client {
    return &Client{
        hub:    hub,
        conn:   conn,
        send:   make(chan *GameEvent, 256),
        RoomID: roomID,
        ID:     id,
    }
}

func (c *Client) SetMessageHandler(handler MessageHandler) {
    c.messageHandler = handler
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
    defer func() {
        c.hub.Unregister <- c
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }

        // Handle message using custom handler if set
        if c.messageHandler != nil {
            if err := c.messageHandler(c, message); err != nil {
                log.Printf("error handling message: %v", err)
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
                // Hub closed the channel
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            err := c.conn.WriteJSON(event)
            if err != nil {
                return
            }

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}