// internal/handlers/game_handler.go

package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/rohan03122001/quizzing/internal/service"
	"github.com/rohan03122001/quizzing/internal/websockets"
)

// GameHandler manages all game-related WebSocket events
type GameHandler struct {
    gameService *service.GameService
    roomService *service.RoomService
    userService *service.UserService
    hub         *websockets.Hub
}

// Message types
const (
    EventJoinRoom     = "join_room"
    EventLeaveRoom    = "leave_room"
    EventStartGame    = "start_game"
    EventSubmitAnswer = "submit_answer"
    EventEndRound     = "end_round"
)

// Data structures for messages
type JoinRoomData struct {
    RoomCode string `json:"room_code"`
    Username string `json:"username"`
}

type SubmitAnswerData struct {
    Answer string `json:"answer"`
}

func NewGameHandler(
    gameService *service.GameService,
    roomService *service.RoomService,
    userService *service.UserService,
    hub *websockets.Hub,
) *GameHandler {
    return &GameHandler{
        gameService: gameService,
        roomService: roomService,
        userService: userService,
        hub:         hub,
    }
}

// HandleMessage routes WebSocket messages to appropriate handlers
func (h *GameHandler) HandleMessage(client *websockets.Client, message []byte) error {
    var event struct {
        Type string          `json:"type"`
        Data json.RawMessage `json:"data"`
    }

    if err := json.Unmarshal(message, &event); err != nil {
        return h.sendErrorToClient(client, "Invalid message format")
    }

    switch event.Type {
    case EventJoinRoom:
        return h.handleJoinRoom(client, event.Data)
    case EventStartGame:
        return h.handleStartGame(client)
    case EventSubmitAnswer:
        return h.handleSubmitAnswer(client, event.Data)
    default:
        return h.sendErrorToClient(client, "Unknown event type")
    }
}

func (h *GameHandler) handleJoinRoom(client *websockets.Client, data json.RawMessage) error {
    var joinData JoinRoomData
    if err := json.Unmarshal(data, &joinData); err != nil {
        return h.sendErrorToClient(client, "Invalid join data format")
    }

    // Join room
    room, err := h.roomService.JoinRoom(joinData.RoomCode, client.ID)
    if err != nil {
        return h.sendErrorToClient(client, err.Error())
    }

    // Update client info
    client.RoomID = room.ID.String()
    client.Username = joinData.Username

    // Notify room
    h.hub.BroadcastToRoom(room.ID.String(), websockets.GameEvent{
        Type: "player_joined",
        Data: map[string]interface{}{
            "player_id": client.ID,
            "username": joinData.Username,
        },
    })

    // Send room state to new player
    return h.hub.SendToClient(client, websockets.GameEvent{
        Type: "room_joined",
        Data: map[string]interface{}{
            "room_code": room.Code,
            "players": h.hub.GetPlayersInRoom(room.ID.String()),
            "settings": map[string]interface{}{
                "max_players": room.MaxPlayers,
                "round_time": room.RoundTime,
                "max_rounds": room.MaxRounds,
            },
        },
    })
}

func (h *GameHandler) handleStartGame(client *websockets.Client) error {
    if err := h.gameService.InitializeGame(client.RoomID); err != nil {
        return h.sendErrorToClient(client, err.Error())
    }

    question, err := h.gameService.StartRound(client.RoomID)
    if err != nil {
        return h.sendErrorToClient(client, err.Error())
    }

    h.hub.BroadcastToRoom(client.RoomID, websockets.GameEvent{
        Type: "round_started",
        Data: map[string]interface{}{
            "question": question,
            "start_time": time.Now(),
        },
    })

    return nil
}

func (h *GameHandler) handleSubmitAnswer(client *websockets.Client, data json.RawMessage) error {
    var answerData SubmitAnswerData
    if err := json.Unmarshal(data, &answerData); err != nil {
        return h.sendErrorToClient(client, "Invalid answer format")
    }

    result, err := h.gameService.ProcessAnswer(client.RoomID, client.ID, answerData.Answer)
    if err != nil {
        return h.sendErrorToClient(client, err.Error())
    }

    // Send result to the answering player
    if err := h.hub.SendToClient(client, websockets.GameEvent{
        Type: "answer_result",
        Data: result,
    }); err != nil {
        return err
    }

    if h.gameService.ShouldEndRound(client.RoomID) {
        return h.endRound(client.RoomID)
    }

    return nil
}

func (h *GameHandler) endRound(roomID string) error {
    if err := h.gameService.EndRound(roomID); err != nil {
        log.Printf("Error ending round: %v", err)
        return err
    }

    if h.gameService.ShouldEndGame(roomID) {
        h.hub.BroadcastToRoom(roomID, websockets.GameEvent{
            Type: "game_end",
        })
        return nil
    }

    // Start next round after delay
    time.AfterFunc(5*time.Second, func() {
        question, err := h.gameService.StartRound(roomID)
        if err != nil {
            log.Printf("Error starting next round: %v", err)
            return
        }

        h.hub.BroadcastToRoom(roomID, websockets.GameEvent{
            Type: "round_started",
            Data: map[string]interface{}{
                "question": question,
                "start_time": time.Now(),
            },
        })
    })

    return nil
}

func (h *GameHandler) sendErrorToClient(client *websockets.Client, message string) error {
    return h.hub.SendToClient(client, websockets.GameEvent{
        Type: "error",
        Data: map[string]string{
            "message": message,
        },
    })
}