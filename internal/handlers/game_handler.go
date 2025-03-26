// internal/handlers/game_handler.go

package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/rohan03122001/quizzing/internal/models"
	"github.com/rohan03122001/quizzing/internal/service"
	"github.com/rohan03122001/quizzing/internal/websocket"
)

// Message types
const (
    EventJoinRoom     = "join_room"
    EventLeaveRoom    = "leave_room"
    EventStartGame    = "start_game"
    EventSubmitAnswer = "submit_answer"
    EventPlayAgain    = "play_again"
    EventReconnect    = "reconnect"
)

// Request structures
type JoinRoomData struct {
    RoomCode string `json:"room_code"`
    Username string `json:"username"`
}

type SubmitAnswerData struct {
    Answer string `json:"answer"`
}

type GameSettings struct {
    MaxRounds  int `json:"max_rounds"`
    RoundTime  int `json:"round_time"`
}

type ReconnectData struct {
    RoomCode  string `json:"room_code"`
    PlayerID  string `json:"player_id"`
    Username  string `json:"username"`
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
    case EventLeaveRoom:
        return h.handleLeaveRoom(client)
    case EventStartGame:
        return h.handleStartGame(client)
    case EventSubmitAnswer:
        return h.handleSubmitAnswer(client, event.Data)
    case EventPlayAgain:
        return h.handlePlayAgain(client, event.Data)
    case EventReconnect:
        return h.handleReconnect(client, event.Data)
    default:
        return h.sendError(client, "Unknown event type")
    }
}

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
    client.RoomID = room.Code
    client.Username = joinData.Username
    client.Status = "active"

    // Register client with hub
    h.hub.Register <- client

    // Get all players in room (including disconnected)
    allPlayers := h.hub.GetAllPlayersInRoom(room.Code)
    activePlayers := h.hub.GetPlayersInRoom(room.Code)

    // Notify room about new player
    h.hub.BroadcastToRoom(room.Code, websocket.GameEvent{
        Type: "player_joined",
        Data: map[string]interface{}{
            "player_id": client.ID,
            "username": joinData.Username,
            "total_players": len(activePlayers),
        },
    })

    // Send room state to new player
    return h.hub.SendToClient(client, websocket.GameEvent{
        Type: "room_joined",
        Data: map[string]interface{}{
            "room_code": room.Code,
            "players": allPlayers, // Send all players with their status
            "active_players": activePlayers, // For backward compatibility
            "settings": map[string]interface{}{
                "max_players": room.MaxPlayers,
                "round_time": room.RoundTime,
                "max_rounds": room.MaxRounds,
            },
        },
    })
}

// Add handler for explicit leave room requests
func (h *GameHandler) handleLeaveRoom(client *websocket.Client) error {
    if client.RoomID == "" {
        return h.sendError(client, "Not in a room")
    }
    
    roomCode := client.RoomID
    
    // Mark client as disconnected
    client.Status = "disconnected"
    
    // Notify other players
    h.hub.BroadcastToRoom(roomCode, websocket.GameEvent{
        Type: "player_left",
        Data: map[string]interface{}{
            "player_id": client.ID,
            "username": client.Username,
        },
    })
    
    // Unregister from hub
    h.hub.Unregister <- client
    
    log.Printf("Player %s explicitly left room %s", client.ID, roomCode)
    return nil
}

func (h *GameHandler) handleStartGame(client *websocket.Client) error {
    log.Printf("Start game request from client %s in room %s", client.ID, client.RoomID)

    // Check if we have enough active players
    activePlayerCount := h.hub.GetActivePlayerCount(client.RoomID)
    if activePlayerCount < 2 {
        return h.sendError(client, "Need at least 2 active players to start")
    }

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
            "active_players": activePlayerCount,
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

        // Get active player count
        activePlayerCount := h.hub.GetActivePlayerCount(roomID)

        h.hub.BroadcastToRoom(roomID, websocket.GameEvent{
            Type: "round_started",
            Data: map[string]interface{}{
                "question": question,
                "active_players": activePlayerCount,
            },
        })
    }()

    return nil
}

func (h *GameHandler) handlePlayAgain(client *websocket.Client, data json.RawMessage) error {
    var settings GameSettings
    if err := json.Unmarshal(data, &settings); err != nil {
        return h.sendError(client, "Invalid settings format")
    }

    // Validate settings
    if settings.MaxRounds <= 0 {
        settings.MaxRounds = 5 // Default
    }
    if settings.RoundTime <= 0 {
        settings.RoundTime = 30 // Default
    }

    // Check active player count
    activePlayerCount := h.hub.GetActivePlayerCount(client.RoomID)
    if activePlayerCount < 2 {
        return h.sendError(client, "Need at least 2 active players to restart game")
    }

    // Restart game
    if err := h.gameService.RestartGame(client.RoomID, &models.GameSettings{
        MaxRounds: settings.MaxRounds,
        RoundTime: settings.RoundTime,
    }); err != nil {
        return h.sendError(client, err.Error())
    }

    return nil
}

func (h *GameHandler) handleReconnect(client *websocket.Client, data json.RawMessage) error {
    var reconnectData ReconnectData
    if err := json.Unmarshal(data, &reconnectData); err != nil {
        return h.sendError(client, "Invalid reconnect data format")
    }
    
    // Verify room exists and is active
    room, err := h.roomService.GetRoom(reconnectData.RoomCode)
    if err != nil {
        return h.sendError(client, "Room not found")
    }

    // Find existing client
    existingClient := h.hub.FindClientByID(reconnectData.RoomCode, reconnectData.PlayerID)
    
    if existingClient == nil {
        // Player wasn't in this room
        return h.sendError(client, "Player not found in this room")
    }
    
    if existingClient.Status == "active" {
        // Someone is already connected with this ID
        return h.sendError(client, "Player already active in this room")
    }
    
    // Update client info from existing client
    client.ID = existingClient.ID
    client.RoomID = room.Code
    client.Username = existingClient.Username
    client.Status = "active"
    
    // Remove old client
    h.hub.RemoveClient(existingClient)
    
    // Register new client in place of old one
    h.hub.Register <- client

    // Get game state
    gameState, err := h.gameService.GetGameState(room.Code, client.ID)
    if err != nil {
        return h.sendError(client, err.Error())
    }

    // Notify other players about reconnection
    h.hub.BroadcastToRoom(room.Code, websocket.GameEvent{
        Type: "player_reconnected",
        Data: map[string]interface{}{
            "player_id": client.ID,
            "username": client.Username,
        },
    })

    // Send current game state to reconnected player
    return h.hub.SendToClient(client, websocket.GameEvent{
        Type: "reconnected",
        Data: gameState,
    })
}

func (h *GameHandler) sendError(client *websocket.Client, message string) error {
    return h.hub.SendToClient(client, websocket.GameEvent{
        Type: "error",
        Data: map[string]string{
            "message": message,
        },
    })
}