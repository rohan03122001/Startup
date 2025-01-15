// internal/handlers/game_handler.go

package handlers

import (
	"encoding/json"

	"github.com/rohan03122001/quizzing/internal/service"
	"github.com/rohan03122001/quizzing/internal/websocket"
)

type GameHandler struct {
    gameService *service.GameService
    roomService *service.RoomService
    hub         *websocket.Hub
}

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
    // Parse message type
    var event struct {
        Type string          `json:"type"`
        Data json.RawMessage `json:"data"`
    }

    if err := json.Unmarshal(message, &event); err != nil {
        return h.sendErrorToClient(client, "Invalid message format")
    }

    switch event.Type {
    case "join_room":
        return h.handleJoinRoom(client, event.Data)
    case "start_game":
        return h.handleStartGame(client)
    case "submit_answer":
        return h.handleSubmitAnswer(client, event.Data)
    default:
        return h.sendErrorToClient(client, "Unknown event type")
    }
}

func (h *GameHandler) handleJoinRoom(client *websocket.Client, data json.RawMessage) error {
    var joinData JoinRoomData
    if err := json.Unmarshal(data, &joinData); err != nil {
        return h.sendErrorToClient(client, "Invalid join data format")
    }

    // Join room
    room, err := h.roomService.JoinRoom(joinData.RoomCode)
    if err != nil {
        return h.sendErrorToClient(client, err.Error())
    }

    // Update client info
    client.RoomID = room.ID.String()
    client.Username = joinData.Username

    // Notify room
    h.hub.BroadcastToRoom(room.ID.String(), websocket.GameEvent{
        Type: "player_joined",
        Data: map[string]interface{}{
            "player_id": client.ID,
            "username": joinData.Username,
        },
    })

    return nil
}

func (h *GameHandler) handleStartGame(client *websocket.Client) error {
    if err := h.roomService.StartGame(client.RoomID); err != nil {
        return h.sendErrorToClient(client, err.Error())
    }

    // Start first round
    question, err := h.gameService.StartRound(client.RoomID)
    if err != nil {
        return h.sendErrorToClient(client, err.Error())
    }

    // Send question to room
    h.hub.BroadcastToRoom(client.RoomID, websocket.GameEvent{
        Type: "new_question",
        Data: question,
    })

    return nil
}

func (h *GameHandler) handleSubmitAnswer(client *websocket.Client, data json.RawMessage) error {
    var answerData SubmitAnswerData
    if err := json.Unmarshal(data, &answerData); err != nil {
        return h.sendErrorToClient(client, "Invalid answer format")
    }

    result, err := h.gameService.ProcessAnswer(client.RoomID, client.ID, answerData.Answer)
    if err != nil {
        return h.sendErrorToClient(client, err.Error())
    }

    // Send result to player
    if err := h.hub.SendToClient(client, websocket.GameEvent{
        Type: "answer_result",
        Data: result,
    }); err != nil {
        return err
    }

    // Check if round should end
    if h.gameService.ShouldEndRound(client.RoomID) {
        if err := h.gameService.EndRound(client.RoomID); err != nil {
            return err
        }
    }

    return nil
}

func (h *GameHandler) sendErrorToClient(client *websocket.Client, message string) error {
    return h.hub.SendToClient(client, websocket.GameEvent{
        Type: "error",
        Data: map[string]string{
            "message": message,
        },
    })
}