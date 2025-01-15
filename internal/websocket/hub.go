package websocket

import (
	"log"
	"sync"
)

type Hub struct {
    // Registered clients mapped by room
    rooms map[string]map[string]*Client

    // Protects rooms map
    mu sync.RWMutex

    // Register requests from clients
    Register chan *Client

    // Unregister requests from clients
    Unregister chan *Client

    // Broadcast messages
    Broadcast chan *GameEvent
}

func NewHub() *Hub {
    return &Hub{
        rooms:      make(map[string]map[string]*Client),
        Register:   make(chan *Client),
        Unregister: make(chan *Client),
        Broadcast:  make(chan *GameEvent, 256),
    }
}

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

    // Initialize room if it doesn't exist
    if _, exists := h.rooms[client.RoomID]; !exists {
        h.rooms[client.RoomID] = make(map[string]*Client)
    }

    // Add client to room
    h.rooms[client.RoomID][client.ID] = client
    
    playerCount := len(h.rooms[client.RoomID])
    log.Printf("Client %s registered in room %s (total players: %d)", client.ID, client.RoomID, playerCount)
}

func (h *Hub) handleUnregister(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if room, exists := h.rooms[client.RoomID]; exists {
        // Remove client from room
        delete(room, client.ID)
        close(client.send)

        playerCount := len(room)
        log.Printf("Client %s unregistered from room %s (remaining players: %d)", 
            client.ID, client.RoomID, playerCount)

        // Remove room if empty
        if playerCount == 0 {
            delete(h.rooms, client.RoomID)
            log.Printf("Room %s removed (empty)", client.RoomID)
        }
    }
}

// GetPlayerCount returns number of players in a room
func (h *Hub) GetPlayerCount(roomID string) int {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if room, exists := h.rooms[roomID]; exists {
        count := len(room)
        log.Printf("Room %s has %d players", roomID, count)
        return count
    }
    
    log.Printf("Room %s not found in hub", roomID)
    return 0
}

// GetPlayersInRoom returns all players in a room
func (h *Hub) GetPlayersInRoom(roomID string) []map[string]string {
    h.mu.RLock()
    defer h.mu.RUnlock()

    var players []map[string]string
    if room, exists := h.rooms[roomID]; exists {
        for _, client := range room {
            players = append(players, map[string]string{
                "id": client.ID,
                "username": client.Username,
            })
        }
        log.Printf("Found %d players in room %s", len(players), roomID)
    } else {
        log.Printf("No players found for room %s", roomID)
    }

    return players
}

func (h *Hub) BroadcastToRoom(roomID string, event GameEvent) {
    event.RoomID = roomID
    h.Broadcast <- &event
}

func (h *Hub) handleBroadcast(event *GameEvent) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if room, exists := h.rooms[event.RoomID]; exists {
        log.Printf("Broadcasting message type %s to %d players in room %s", 
            event.Type, len(room), event.RoomID)

        for _, client := range room {
            select {
            case client.send <- event:
                // Message sent successfully
            default:
                close(client.send)
                log.Printf("Failed to send message to client %s (buffer full)", client.ID)
            }
        }
    }
}




type GameEvent struct {
    Type    string      `json:"type"`
    RoomID  string      `json:"room_id,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// internal/websocket/hub.go (continued)

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(client *Client, event GameEvent) error {
    select {
    case client.send <- &event:
        return nil
    default:
        close(client.send)
        h.Unregister <- client
        return nil
    }
}


