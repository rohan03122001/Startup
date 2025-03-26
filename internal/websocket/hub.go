// internal/websocket/hub.go

package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
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

func (h *Hub) handleRegister(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    // Initialize room if it doesn't exist
    if _, exists := h.rooms[client.RoomID]; !exists {
        h.rooms[client.RoomID] = make(map[string]*Client)
    }

    // Add client to room
    h.rooms[client.RoomID][client.ID] = client
    client.Status = "active" // Ensure status is set to active
    
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
            // Only include active players in this method for backward compatibility
            if client.Status == "active" {
                players = append(players, map[string]string{
                    "id": client.ID,
                    "username": client.Username,
                })
            }
        }
        log.Printf("Found %d active players in room %s", len(players), roomCode)
    } else {
        log.Printf("Room %s not found in hub", roomCode)
    }

    return players
}

// GetAllPlayersInRoom returns all players including disconnected ones
func (h *Hub) GetAllPlayersInRoom(roomCode string) []map[string]string {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    var players []map[string]string
    
    if room, exists := h.rooms[roomCode]; exists {
        for _, client := range room {
            players = append(players, map[string]string{
                "id": client.ID,
                "username": client.Username,
                "status": client.Status,
            })
        }
        log.Printf("Found %d total players (active and disconnected) in room %s", 
            len(players), roomCode)
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
        log.Printf("Room %s has %d total players", roomCode, count)
        return count
    }
    
    log.Printf("Room %s not found in hub or empty", roomCode)
    return 0
}

// GetActivePlayerCount returns only active (connected) players
func (h *Hub) GetActivePlayerCount(roomCode string) int {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    activeCount := 0
    
    if room, exists := h.rooms[roomCode]; exists {
        for _, client := range room {
            if client.Status == "active" {
                activeCount++
            }
        }
        log.Printf("Room %s has %d active players", roomCode, activeCount)
    }
    
    return activeCount
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

// Handle client disconnection with grace period
func (h *Hub) handleDisconnection(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if room, exists := h.rooms[client.RoomID]; exists {
        if _, ok := room[client.ID]; ok {
            // Mark client as disconnected instead of removing immediately
            client.Status = "disconnected"
            
            // Generate reconnection token and store it
            client.ReconnectToken = uuid.New().String()
            
            log.Printf("Client %s disconnected from room %s (marked as disconnected)", 
                client.ID, client.RoomID)
                
            // Broadcast disconnection to room
            go h.BroadcastToRoom(client.RoomID, GameEvent{
                Type: "player_disconnected",
                Data: map[string]interface{}{
                    "player_id": client.ID,
                    "username": client.Username,
                },
            })
            
            // Start cleanup timer for this client
            go h.scheduleClientCleanup(client)
        }
    }
}

// Schedule client cleanup after disconnect timeout
func (h *Hub) scheduleClientCleanup(client *Client) {
    disconnectTimeout := 5 * time.Minute
    
    time.Sleep(disconnectTimeout)
    
    h.mu.Lock()
    defer h.mu.Unlock()
    
    // Check if client is still disconnected after timeout
    if room, exists := h.rooms[client.RoomID]; exists {
        if c, ok := room[client.ID]; ok && c.Status == "disconnected" {
            // Now fully remove the client
            delete(room, client.ID)
            close(client.send)
            
            log.Printf("Removed disconnected client %s after timeout", client.ID)
            
            // Remove room if empty
            if len(room) == 0 {
                delete(h.rooms, client.RoomID)
                log.Printf("Room %s removed (empty after disconnect timeout)", client.RoomID)
            }
        }
    }
}

func (h *Hub) handleUnregister(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if room, exists := h.rooms[client.RoomID]; exists {
        if _, ok := room[client.ID]; ok {
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
}

func (h *Hub) handleBroadcast(event *GameEvent) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if room, exists := h.rooms[event.RoomID]; exists {
        activeCount := 0
        for _, client := range room {
            if client.Status == "active" {
                activeCount++
            }
        }
        
        log.Printf("Broadcasting %s event to room %s (%d active players out of %d total)", 
            event.Type, event.RoomID, activeCount, len(room))

        for _, client := range room {
            // Only send to active clients
            if client.Status == "active" {
                select {
                case client.send <- event:
                    // Message sent successfully
                default:
                    // Client's buffer is full, mark as disconnected
                    client.Status = "disconnected"
                    go h.scheduleClientCleanup(client)
                    log.Printf("Client %s buffer full, marked as disconnected", client.ID)
                }
            }
        }
    }
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(client *Client, event GameEvent) error {
    if client.Status != "active" {
        return nil // Don't send to inactive clients
    }
    
    select {
    case client.send <- &event:
        return nil
    default:
        return nil
    }
}

// FindClientByID finds a client in a room by ID
func (h *Hub) FindClientByID(roomCode string, clientID string) *Client {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    if room, exists := h.rooms[roomCode]; exists {
        if client, ok := room[clientID]; ok {
            return client
        }
    }
    
    return nil
}

// RemoveClient removes a client from the hub
func (h *Hub) RemoveClient(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if room, exists := h.rooms[client.RoomID]; exists {
        if _, ok := room[client.ID]; ok {
            delete(room, client.ID)
            close(client.send)
            
            // Remove room if empty
            if len(room) == 0 {
                delete(h.rooms, client.RoomID)
            }
        }
    }
}