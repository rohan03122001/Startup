// internal/handlers/game_handler.go

package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/rohan03122001/quizzing/internal/service"
	"github.com/rohan03122001/quizzing/internal/websocket"
)

// Message types
const (
    EventJoinRoom     = "join_room"
    EventLeaveRoom    = "leave_room"
    EventStartGame    = "start_game"
    EventSubmitAnswer = "submit_answer"
)

// Request structures
type JoinRoomData struct {
    RoomCode string `json:"room_code"`
    Username string `json:"username"`
}

type SubmitAnswerData struct {
    Answer string `json:"answer"`
}

type GameHandler struct {
    gameService *service.GameService
    roomService *service.RoomService
    hub         *websocket.Hub
}

func NewGameHandler(
    gameService *service.GameService,
    roomService *service.RoomService,
    hub *websocket.Hub,
) *GameHandler {
    return &GameHandler{
        gameService: gameService,
        roomService: roomService,
        hub:         hub,
    }
}

// HandleMessage processes incoming WebSocket messages
func (h *GameHandler) HandleMessage(client *websocket.Client, message []byte) error {
    var event struct {
        Type string          `json:"type"`
        Data json.RawMessage `json:"data"`
    }

    if err := json.Unmarshal(message, &event); err != nil {
        return h.sendError(client, "Invalid message format")
    }

    log.Printf("Received message type: %s from client: %s", event.Type, client.ID)

    switch event.Type {
    case EventJoinRoom:
        return h.handleJoinRoom(client, event.Data)
    case EventStartGame:
        return h.handleStartGame(client)
    case EventSubmitAnswer:
        return h.handleSubmitAnswer(client, event.Data)
    default:
        return h.sendError(client, "Unknown event type")
    }
}

// internal/handlers/game_handler.go

func (h *GameHandler) handleJoinRoom(client *websocket.Client, data json.RawMessage) error {
    var joinData JoinRoomData
    if err := json.Unmarshal(data, &joinData); err != nil {
        return h.sendError(client, "Invalid join data format")
    }

    // Join room
    room, err := h.roomService.JoinRoom(joinData.RoomCode, client.ID)
    if err != nil {
        return h.sendError(client, err.Error())
    }

    // Update client info
    client.RoomID = room.Code  // Changed from ID to Code
    client.Username = joinData.Username

    // Register client with hub
    h.hub.Register <- client

    // Get current players after registration
    players := h.hub.GetPlayersInRoom(room.Code)  // Changed from ID to Code

    // Notify room about new player
    h.hub.BroadcastToRoom(room.Code, websocket.GameEvent{
        Type: "player_joined",
        Data: map[string]interface{}{
            "player_id": client.ID,
            "username": joinData.Username,
            "total_players": len(players) + 1,
        },
    })

    // Send room state to new player
    return h.hub.SendToClient(client, websocket.GameEvent{
        Type: "room_joined",
        Data: map[string]interface{}{
            "room_code": room.Code,
            "players": players,
            "settings": map[string]interface{}{
                "max_players": room.MaxPlayers,
                "round_time": room.RoundTime,
                "max_rounds": room.MaxRounds,
            },
        },
    })
}

func (h *GameHandler) handleStartGame(client *websocket.Client) error {
    log.Printf("Start game request from client %s in room %s", client.ID, client.RoomID)

    // Start the game using room code
    if err := h.roomService.StartGame(client.RoomID); err != nil {
        return h.sendError(client, err.Error())
    }

    // Start first round
    question, err := h.gameService.StartRound(client.RoomID)
    if err != nil {
        return h.sendError(client, err.Error())
    }

    // Broadcast question to room
    h.hub.BroadcastToRoom(client.RoomID, websocket.GameEvent{
        Type: "round_started",
        Data: map[string]interface{}{
            "question": question,
            "round_number": 1,
        },
    })

    return nil
}

func (h *GameHandler) handleSubmitAnswer(client *websocket.Client, data json.RawMessage) error {
    var answerData SubmitAnswerData
    if err := json.Unmarshal(data, &answerData); err != nil {
        return h.sendError(client, "Invalid answer format")
    }

    // Process answer
    result, err := h.gameService.ProcessAnswer(client.RoomID, client.ID, answerData.Answer)
    if err != nil {
        return h.sendError(client, err.Error())
    }

    // Send result to the player
    err = h.hub.SendToClient(client, websocket.GameEvent{
        Type: "answer_result",
        Data: result,
    })
    if err != nil {
        return err
    }

    // Check if round should end
    if h.gameService.ShouldEndRound(client.RoomID) {
        return h.endRound(client.RoomID)
    }

    return nil
}

func (h *GameHandler) endRound(roomID string) error {
    if err := h.gameService.EndRound(roomID); err != nil {
        return err
    }

    // Start next round after delay
    go func() {
        time.Sleep(5 * time.Second)
        question, err := h.gameService.StartRound(roomID)
        if err != nil {
            log.Printf("Error starting next round: %v", err)
            return
        }

        h.hub.BroadcastToRoom(roomID, websocket.GameEvent{
            Type: "round_started",
            Data: map[string]interface{}{
                "question": question,
            },
        })
    }()

    return nil
}

func (h *GameHandler) sendError(client *websocket.Client, message string) error {
    return h.hub.SendToClient(client, websocket.GameEvent{
        Type: "error",
        Data: map[string]string{
            "message": message,
        },
    })
}