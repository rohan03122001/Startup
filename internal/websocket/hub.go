// internal/websocket/hub.go

package websocket

import (
	"log"
	"sync"
)

// RoomService interface represents minimal room service methods needed by hub
type RoomService interface {
	UpdateRoomActivity(roomCode string) error
}

// Hub maintains the set of active clients and broadcasts messages
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

    // Room service for updating room activity
    roomService RoomService
}

// GameEvent represents a game-related message
type GameEvent struct {
    Type    string      `json:"type"`
    RoomID  string      `json:"room_id,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// NewHub creates a new Hub for WebSocket communication
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan *GameEvent, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		rooms:      make(map[string]map[string]*Client),
		mu:         sync.RWMutex{},
	}
}

// SetRoomService sets the room service for the hub
func (h *Hub) SetRoomService(service RoomService) {
	h.roomService = service
}

// Run starts the hub's main event loop
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

// internal/websocket/hub.go

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
    log.Printf("Client %s registered in room %s (total players: %d)", 
        client.ID, client.RoomID, playerCount)
}

func (h *Hub) GetPlayersInRoom(roomCode string) []map[string]string {
    h.mu.RLock()
    defer h.mu.RUnlock()

    var players []map[string]string
    if room, exists := h.rooms[roomCode]; exists {
        for _, client := range room {
            players = append(players, map[string]string{
                "id": client.ID,
                "username": client.Username,
            })
        }
        log.Printf("Found %d players in room %s", len(players), roomCode)
    } else {
        log.Printf("Room %s not found in hub", roomCode)
    }

    return players
}

func (h *Hub) GetPlayerCount(roomCode string) int {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if room, exists := h.rooms[roomCode]; exists {
        count := len(room)
        log.Printf("Room %s has %d players", roomCode, count)
        return count
    }
    
    log.Printf("Room %s not found in hub or empty", roomCode)
    return 0
}

// BroadcastToRoom updated for better logging
func (h *Hub) BroadcastToRoom(roomCode string, event GameEvent) {
    event.RoomID = roomCode
    log.Printf("Broadcasting %s event to room %s", event.Type, roomCode)
    
    // Update room last activity if needed
    if h.roomService != nil {
        h.roomService.UpdateRoomActivity(roomCode)
    }
    
    h.Broadcast <- &event
}

func (h *Hub) handleUnregister(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if room, exists := h.rooms[client.RoomID]; exists {
        if _, ok := room[client.ID]; ok {
            delete(room, client.ID)
            close(client.send)
            playerCount := len(room)
            
            log.Printf("Client %s left room %s (remaining players: %d)", 
                client.ID, client.RoomID, playerCount)

            // Remove room if empty
            if playerCount == 0 {
                delete(h.rooms, client.RoomID)
                log.Printf("Room %s removed (empty)", client.RoomID)
            }
        }
    }
}

func (h *Hub) handleBroadcast(event *GameEvent) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if room, exists := h.rooms[event.RoomID]; exists {
        log.Printf("Broadcasting %s event to room %s (%d players)", 
            event.Type, event.RoomID, len(room))

        for _, client := range room {
            select {
            case client.send <- event:
                // Message sent successfully
            default:
                // Client's buffer is full, remove them
                close(client.send)
                delete(room, client.ID)
                log.Printf("Removed client %s (buffer full)", client.ID)
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
        return nil
    }
}