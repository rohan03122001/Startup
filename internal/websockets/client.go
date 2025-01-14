// internal/websocket/client.go

package websockets

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

    // Send pings to peer with this period
    pingPeriod = (pongWait * 9) / 10

    // Maximum message size allowed from peer
    maxMessageSize = 512
)

// Client represents a connected websocket client
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
}

// NewClient creates a new client instance
func NewClient(hub *Hub, conn *websocket.Conn, roomID, clientID string) *Client {
    return &Client{
        hub:    hub,
        conn:   conn,
        send:   make(chan *GameEvent, 256),
        RoomID: roomID,
        ID:     clientID,
    }
}

// SetMessageHandler sets the function to handle incoming messages
func (c *Client) SetMessageHandler(handler func(*Client, []byte) error) {
    c.messageHandler = handler
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
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
                log.Printf("error: %v", err)
            }
            break
        }

        // Handle message using custom handler if set
        if c.messageHandler != nil {
            if err := c.messageHandler(c, message); err != nil {
                log.Printf("error handling message: %v", err)
            }
            continue
        }

        // Default behavior: broadcast to room
        c.hub.events <- &GameEvent{
            Type:   "message",
            RoomID: c.RoomID,
            Data:   string(message),
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

            err := c.conn.WriteJSON(event)
            if err != nil {
                log.Printf("error writing message: %v", err)
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

// SendMessage sends a message to this client
func (c *Client) SendMessage(event *GameEvent) error {
    select {
    case c.send <- event:
        return nil
    default:
        return nil
    }
}

// Close closes the client connection
func (c *Client) Close() {
    c.conn.Close()
}






//---------------------------------------------------------
// package websockets

// import (
// 	"log"
// 	"time"

// 	"github.com/gorilla/websocket"
// )

// const(
// 	writeWait = 10 *time.Second
// 	pongWait = 60 *time.Second
// 	pingPeriod = (pongWait*9)/10
// 	maxMessageSize = 512
// )

// type Client struct {
// 	hub *Hub

// 	// WebSocket connection
// 	conn *websocket.Conn

// 	// Buffered channel of outbound messages
// 	send chan *GameEvent

// 	// Room this client is in
// 	RoomID string

// 	// Unique client identifier
// 	ID string
// }

// func NewClient(hub *Hub, conn *websocket.Conn, roomID string, id string) *Client{
// 	return &Client{
// 		hub: hub,
// 		conn: conn,
// 		send: make(chan *GameEvent, 256),
// 		RoomID: roomID,
// 		ID: id,
// 	}
// }


// func(c *Client)ReadPump(){
// 	defer func ()  {
// 		c.hub.Unregister <- c
// 		c.conn.Close()
// 	}()

// 	c.conn.SetReadLimit(maxMessageSize)
// 	c.conn.SetReadDeadline(time.Now().Add(pongWait))
// 	c.conn.SetPongHandler(func(string) error {
// 		c.conn.SetReadDeadline(time.Now().Add(pongWait))
// 		return nil
// 	})


// 	for{
// 		var gameEvent GameEvent
// 		err := c.conn.ReadJSON(&gameEvent)

// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure){
// 				log.Printf("error: %v", err)
// 			}
// 			break
// 		}

// 		gameEvent.RoomID = c.RoomID

// 		c.hub.Events <- &gameEvent
// 	}
// }


// func (c *Client) WritePump(){
// 	ticker := time.NewTicker(pingPeriod)

// 	defer func ()  {
// 		ticker.Stop()
// 		c.conn.Close()
// 	}()

// 	for{
// 		select{
// 		case event, ok := <-c.send:
// 			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
// 			if !ok{
// 				//hub channel closed
// 				c.conn.WriteMessage(websocket.CloseMessage,[]byte{})
// 				return 
// 			}

// 			err:= c.conn.WriteJSON(event)
// 			if err != nil{
// 				return
// 			}
			
// 		case <-ticker.C:
// 			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
// 			if err:= c.conn.WriteMessage(websocket.PingMessage, nil); err!=nil{
// 				return
// 			}
// 		}
// 	}
// }



// internal/websocket/client.go
