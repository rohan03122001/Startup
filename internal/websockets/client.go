package websockets

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const(
	writeWait = 10 *time.Second
	pongWait = 60 *time.Second
	pingPeriod = (pongWait*9)/10
	maxMessageSize = 512
)

type Client struct {
	hub *Hub

	// WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan *GameEvent

	// Room this client is in
	RoomID string

	// Unique client identifier
	ID string
}

func NewClient(hub *Hub, conn *websocket.Conn, roomID string, id string) *Client{
	return &Client{
		hub: hub,
		conn: conn,
		send: make(chan *GameEvent, 256),
		RoomID: roomID,
		ID: id,
	}
}


func(c *Client)ReadPump(){
	defer func ()  {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})


	for{
		var gameEvent GameEvent
		err := c.conn.ReadJSON(&gameEvent)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure){
				log.Printf("error: %v", err)
			}
			break
		}

		gameEvent.RoomID = c.RoomID

		c.hub.Events <- &gameEvent
	}
}


func (c *Client) WritePump(){
	ticker := time.NewTicker(pingPeriod)

	defer func ()  {
		ticker.Stop()
		c.conn.Close()
	}()

	for{
		select{
		case event, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok{
				//hub channel closed
				c.conn.WriteMessage(websocket.CloseMessage,[]byte{})
				return 
			}

			err:= c.conn.WriteJSON(event)
			if err != nil{
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err:= c.conn.WriteMessage(websocket.PingMessage, nil); err!=nil{
				return
			}
		}
	}
}
