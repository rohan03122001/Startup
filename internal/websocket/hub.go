package websocket

import "sync"

type Hub struct {
	//Map of all active rooms -> map of room1:{string:&client{}}rooms={room1:{rohan}, room2:anant}
	// rooms = {
	// 	"room1": {
	// 		"client1": &Client{
	// 			ID: "client1",
	// 			RoomID: "room1",
	// 			send: make(chan *GameEvent, 10),
	// 			conn: <WebSocket connection for client1>,
	// 			hub: <Hub reference>,
	// 		},
	// 		"client2": &Client{
	// 			ID: "client2",
	// 			RoomID: "room1",
	// 			send: make(chan *GameEvent, 10),
	// 			conn: <WebSocket connection for client2>,
	// 			hub: <Hub reference>,
	// 		},
	// 	},
	// 	"room2": {
	// 		"client3": &Client{
	// 			ID: "client3",
	// 			RoomID: "room2",
	// 			send: make(chan *GameEvent, 10),
	// 			conn: <WebSocket connection for client3>,
	// 			hub: <Hub reference>,
	// 		},
	// 	},
	//}
	rooms map[string]map[string]*Client

	//Protect Access
	mu sync.RWMutex

	//Register
	register chan *Client

	//Unregister
	unregister chan *Client

	//Game Evenet
	events chan *GameEvent
}

type GameEvent struct{
	Type string `json:"type"`
	RoomID string	`json:"room_id"`
	Data	interface{}	`json:"data"`
}

func NewHub() *Hub{
	return &Hub{
		rooms: make(map[string]map[string]*Client),
		register: make(chan *Client),
		unregister: make(chan *Client),
		events: make(chan *GameEvent),
	}
}

func(h *Hub)Run(){
	for{
		select{
		case client := <-h.register:
			h.handleRegister(client)
		case client := <-h.unregister:
			h.handleUnregister(client)
		
		case event := <-h.events:
			h.handleEvent(event)
		}
	}
}


func(h *Hub)handleRegister(client *Client){
	h.mu.Lock()
	defer h.mu.Unlock()

	//initialise if room doesnt exists 
	//we check if the roomID in the client we received already exists in our outer map rooms[roomID] 
	if _ , exists := h.rooms[client.RoomID]; exists{
		h.rooms[client.RoomID] = make(map[string]*Client)
	}


	//its a map of room(string) : client(string) :(client){ which also has a reference to its room}
	//add client
	h.rooms[client.RoomID][client.ID] = client

	//h.roooms["room1"][rohan[1]] = rohan
	//notify others about player joining	except player itself							
	h.broadcastToRoom(client.RoomID, &GameEvent{
		Type: "player_joined",
		RoomID: client.RoomID,
		Data: map[string]string{
			"player_id": client.ID,
		},
	},client.ID)
}



func(h *Hub)handleUnregister(client *Client){
	h.mu.Lock()
	defer h.mu.Unlock()

	//check room exist
	if room, exists := h.rooms[client.RoomID]; exists{
		if _, ok:= room[client.ID]; ok{
			delete(room, client.ID)
			close(client.send)


			if len(room) == 0 {
				delete(h.rooms, client.RoomID)
			} else {
				h.broadcastToRoom(client.RoomID,&GameEvent{
					Type :"player_left",
					RoomID: client.RoomID,
					Data: map[string]string{
						"player_id":client.ID,
					},
				},"")
			}
		}
	}
}


func (h *Hub)handleEvent(event *GameEvent){
	h.mu.Lock()
	defer h.mu.Unlock()

	h.broadcastToRoom(event.RoomID, event, "")
}

func (h *Hub)broadcastToRoom(roomID string, event *GameEvent, excludeClientID string){
	if room, exists := h.rooms[roomID]; exists{
		for clientID, client := range room{
			if clientID != excludeClientID{
				select{
				case client.send <- event:
					//msg sent successfully
				default:
					close(client.send)
					delete(room, clientID)
				}
			}
		}
	}
}


func (h *Hub) BroadbcastQuestion(roomID string, question interface{}){
	h.events <- &GameEvent{
		Type: "new_question",
		RoomID: roomID,
		Data: question,
	}
}

func(h *Hub) BroadcastResult(roomID string, result interface{}){
	h.events <- &GameEvent{
		Type: "round_result",
		RoomID: roomID,
		Data: result,
	}
}
