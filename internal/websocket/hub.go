// internal/websocket/hub.go

package websocket

import (
	"fmt"
	"sync"
)

// Hub manages all active WebSocket connections
type Hub struct {
    // Registered clients mapped by room
    rooms map[string]map[string]*Client

    // Protects rooms map
    mu sync.RWMutex

    // Register requests from clients
    Register chan *Client

    // Unregister requests from clients
    Unregister chan *Client

    // Game events channel
    Broadcast chan *GameEvent
}

// GameEvent represents any game-related event
type GameEvent struct {
    Type    string      `json:"type"`
    RoomID  string      `json:"room_id,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func NewHub() *Hub {
    return &Hub{
        rooms:      make(map[string]map[string]*Client),
        Register:   make(chan *Client),
        Unregister: make(chan *Client),
        Broadcast:  make(chan *GameEvent, 256),
    }
}

// Run starts the hub
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.Register:
            h.handleRegister(client)

        case client := <-h.Unregister:
            h.handleUnregister(client)

        case event := <-h.Broadcast:
            h.handleBroadcast(event)
        }
    }
}

func (h *Hub) handleRegister(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if _, exists := h.rooms[client.RoomID]; !exists {
        h.rooms[client.RoomID] = make(map[string]*Client)
    }
    h.rooms[client.RoomID][client.ID] = client
}

func (h *Hub) handleUnregister(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if room, exists := h.rooms[client.RoomID]; exists {
        if _, ok := room[client.ID]; ok {
            delete(room, client.ID)
            close(client.send)
            
            if len(room) == 0 {
                delete(h.rooms, client.RoomID)
            }
        }
    }
}

func (h *Hub) handleBroadcast(event *GameEvent) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if room, exists := h.rooms[event.RoomID]; exists {
        for _, client := range room {
            select {
            case client.send <- event:
                // Message sent successfully
            default:
                close(client.send)
                go func(c *Client) {
                    h.Unregister <- c
                }(client)
            }
        }
    }
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(client *Client, event GameEvent) error {
    select {
    case client.send <- &event:
        return nil
    default:
        close(client.send)
        go func() {
            h.Unregister <- client
        }()
        return fmt.Errorf("client send buffer full")
    }
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(roomID string, event GameEvent) {
    event.RoomID = roomID
    h.Broadcast <- &event
}

// GetPlayerCount returns number of players in a room
func (h *Hub) GetPlayerCount(roomID string) int {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    if room, exists := h.rooms[roomID]; exists {
        return len(room)
    }
    return 0
}

// GetPlayersInRoom gets all players in a room
func (h *Hub) GetPlayersInRoom(roomID string) []string {
    h.mu.RLock()
    defer h.mu.RUnlock()

    var players []string
    if room, exists := h.rooms[roomID]; exists {
        for clientID := range room {
            players = append(players, clientID)
        }
    }
    return players
}