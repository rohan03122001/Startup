package websocket

import "github.com/gorilla/websocket"

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