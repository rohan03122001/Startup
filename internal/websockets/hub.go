package websockets

// internal/websocket/hub.go

import (
	"errors"
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
    // Registered clients mapped by room
    rooms map[string]map[string]*Client

    // Protects rooms map
    mu sync.RWMutex

    // Register requests from clients
    register chan *Client

    // Unregister requests from clients
    unregister chan *Client

    // Game events channel
    events chan *GameEvent

    // Optional callback when room is empty
    onRoomEmpty func(roomID string)
}

type GameEvent struct {
    Type    string      `json:"type"`
    RoomID  string      `json:"room_id,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func NewHub() *Hub {
    return &Hub{
        rooms:      make(map[string]map[string]*Client),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        events:     make(chan *GameEvent, 256),
    }
}

// Run starts the hub's main loop
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.HandleRegister(client)

        case client := <-h.unregister:
            h.HandleUnregister(client)

        case event := <-h.events:
            h.handleEvent(event)
        }
    }
}

// handleRegister adds a new client to a room
func (h *Hub) HandleRegister(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    // Create room if it doesn't exist
    if _, exists := h.rooms[client.RoomID]; !exists {
        h.rooms[client.RoomID] = make(map[string]*Client)
    }

    // Add client to room
    h.rooms[client.RoomID][client.ID] = client

    log.Printf("Client %s joined room %s", client.ID, client.RoomID)
}

// handleUnregister removes a client
func (h *Hub) HandleUnregister(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if room, exists := h.rooms[client.RoomID]; exists {
        if _, ok := room[client.ID]; ok {
            delete(room, client.ID)
            close(client.send)

            log.Printf("Client %s left room %s", client.ID, client.RoomID)

            // If room is empty
            if len(room) == 0 {
                delete(h.rooms, client.RoomID)
                if h.onRoomEmpty != nil {
                    go h.onRoomEmpty(client.RoomID)
                }
            } else {
                // Notify others about player leaving
                event := &GameEvent{
                    Type: "player_left",
                    Data: map[string]string{
                        "player_id": client.ID,
                        "username": client.Username,
                    },
                }
                h.broadcastToRoom(client.RoomID, event, "")
            }
        }
    }
}

// handleEvent processes and broadcasts game events
func (h *Hub) handleEvent(event *GameEvent) {
    h.mu.RLock()
    if room, exists := h.rooms[event.RoomID]; exists {
        for _, client := range room {
            select {
            case client.send <- event:
                // Message sent successfully
            default:
                // Client's buffer is full
                close(client.send)
                h.unregister <- client
            }
        }
    }
    h.mu.RUnlock()
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(client *Client, event GameEvent) error {
    select {
    case client.send <- &event:
        return nil
    default:
        return errors.New("client send buffer full")
    }
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(roomID string, event GameEvent) {
    h.events <- &event
}

// GetPlayersInRoom gets all players in a room
func (h *Hub) GetPlayersInRoom(roomID string) []map[string]string {
    h.mu.RLock()
    defer h.mu.RUnlock()

    var players []map[string]string
    if room, exists := h.rooms[roomID]; exists {
        for _, client := range room {
            players = append(players, map[string]string{
                "id":       client.ID,
                "username": client.Username,
            })
        }
    }
    return players
}

// GetPlayerCount returns the number of players in a room
func (h *Hub) GetPlayerCount(roomID string) int {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if room, exists := h.rooms[roomID]; exists {
        return len(room)
    }
    return 0
}

// SetOnRoomEmpty sets callback for when a room becomes empty
func (h *Hub) SetOnRoomEmpty(callback func(roomID string)) {
    h.onRoomEmpty = callback
}

// broadcastToRoom sends a message to all clients in a room except one
func (h *Hub) broadcastToRoom(roomID string, event *GameEvent, excludeClientID string) {
    if room, exists := h.rooms[roomID]; exists {
        for clientID, client := range room {
            if clientID != excludeClientID {
                select {
                case client.send <- event:
                    // Message sent successfully
                default:
                    close(client.send)
                    delete(room, clientID)
                }
            }
        }
    }
}



// ----------------------------------------------


// package websockets

// import "sync"

// type Hub struct {
// 	//Map of all active rooms -> map of room1:{string:&client{}}rooms={room1:{rohan}, room2:anant}
// 	// rooms = {
// 	// 	"room1": {
// 	// 		"client1": &Client{
// 	// 			ID: "client1",
// 	// 			RoomID: "room1",
// 	// 			send: make(chan *GameEvent, 10),
// 	// 			conn: <WebSocket connection for client1>,
// 	// 			hub: <Hub reference>,
// 	// 		},
// 	// 		"client2": &Client{
// 	// 			ID: "client2",
// 	// 			RoomID: "room1",
// 	// 			send: make(chan *GameEvent, 10),
// 	// 			conn: <WebSocket connection for client2>,
// 	// 			hub: <Hub reference>,
// 	// 		},
// 	// 	},
// 	// 	"room2": {
// 	// 		"client3": &Client{
// 	// 			ID: "client3",
// 	// 			RoomID: "room2",
// 	// 			send: make(chan *GameEvent, 10),
// 	// 			conn: <WebSocket connection for client3>,
// 	// 			hub: <Hub reference>,
// 	// 		},
// 	// 	},
// 	//}
// 	Rooms map[string]map[string]*Client

// 	//Protect Access
// 	mu sync.RWMutex

// 	//Register
// 	Register chan *Client

// 	//Unregister
// 	Unregister chan *Client

// 	//Game Evenet
// 	Events chan *GameEvent

// 	//callback when room is empty
// 	onRoomEmpty func(roomID string)
// }

// type GameEvent struct{
// 	Type string `json:"type"`
// 	RoomID string	`json:"room_id"`
// 	Data	interface{}	`json:"data"`
	
// }

// func NewHub() *Hub{
// 	return &Hub{
// 		Rooms: make(map[string]map[string]*Client),
// 		Register: make(chan *Client),
// 		Unregister: make(chan *Client),
// 		Events: make(chan *GameEvent),
// 	}
// }

// func(h *Hub)Run(){
// 	for{
// 		select{
// 		case client := <-h.Register:
// 			h.handleRegister(client)
// 		case client := <-h.Unregister:
// 			h.handleUnregister(client)
		
// 		case event := <-h.Events:
// 			h.handleEvent(event)
// 		}
// 	}
// }


// func(h *Hub)handleRegister(client *Client){
// 	h.mu.Lock()
// 	defer h.mu.Unlock()

// 	//initialise if room doesnt exists 
// 	//we check if the roomID in the client we received already exists in our outer map rooms[roomID] 
// 	if _ , exists := h.Rooms[client.RoomID]; exists{
// 		h.Rooms[client.RoomID] = make(map[string]*Client)
// 	}


// 	//its a map of room(string) : client(string) :(client){ which also has a reference to its room}
// 	//add client
// 	h.Rooms[client.RoomID][client.ID] = client

// 	//h.roooms["room1"][rohan[1]] = rohan
// 	//notify others about player joining	except player itself							
// 	h.broadcastToRoom(client.RoomID, &GameEvent{
// 		Type: "player_joined",
// 		RoomID: client.RoomID,
// 		Data: map[string]string{
// 			"player_id": client.ID,
// 		},
// 	},client.ID)
// }



// func(h *Hub)handleUnregister(client *Client){
// 	h.mu.Lock()
// 	defer h.mu.Unlock()

// 	//check room exist
// 	if room, exists := h.Rooms[client.RoomID]; exists{
// 		if _, ok:= room[client.ID]; ok{
// 			delete(room, client.ID)
// 			close(client.send)


// 			if len(room) == 0 {
// 				delete(h.Rooms, client.RoomID)
// 			} else {
// 				h.broadcastToRoom(client.RoomID,&GameEvent{
// 					Type :"player_left",
// 					RoomID: client.RoomID,
// 					Data: map[string]string{
// 						"player_id":client.ID,
// 					},
// 				},"")
// 			}
// 		}
// 	}
// }


// func (h *Hub)handleEvent(event *GameEvent){
// 	h.mu.Lock()
// 	defer h.mu.Unlock()

// 	h.broadcastToRoom(event.RoomID, event, "")
// }

// func (h *Hub)broadcastToRoom(roomID string, event *GameEvent, excludeClientID string){
// 	if room, exists := h.Rooms[roomID]; exists{
// 		for clientID, client := range room{
// 			if clientID != excludeClientID{
// 				select{
// 				case client.send <- event:
// 					//msg sent successfully
// 				default:
// 					close(client.send)
// 					delete(room, clientID)
// 				}
// 			}
// 		}
// 	}
// }


// func (h *Hub) BroadbcastQuestion(roomID string, question interface{}){
// 	h.Events <- &GameEvent{
// 		Type: "new_question",
// 		RoomID: roomID,
// 		Data: question,
// 	}
// }

// func(h *Hub) BroadcastResult(roomID string, result interface{}){
// 	h.Events <- &GameEvent{
// 		Type: "round_result",
// 		RoomID: roomID,
// 		Data: result,
// 	}
// }