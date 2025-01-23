// internal/handlers/http_handler.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohan03122001/quizzing/internal/service"
)

type HTTPHandler struct {
    roomService *service.RoomService
}

func NewHTTPHandler(roomService *service.RoomService) *HTTPHandler {
    return &HTTPHandler{
        roomService: roomService,
    }
}

// CreateRoom handles room creation
func (h *HTTPHandler) CreateRoom(c *gin.Context) {
    room, err := h.roomService.CreateRoom()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, room)
    // c.JSON(http.StatusCreated, gin.H{
    //     "room_code": room.Code,
    //     "max_players": room.MaxPlayers,
    //     "round_time": room.RoundTime,
    // })
}

// GetActiveRooms returns list of rooms waiting for players
func (h *HTTPHandler) GetActiveRooms(c *gin.Context) {
    rooms, err := h.roomService.GetActiveRooms()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, rooms)
}

// Add this new struct
type JoinRoomRequest struct {
    RoomCode string `json:"room_code"`
    Username string `json:"username"`
}

type JoinRoomResponse struct {
    RoomCode    string `json:"room_code"`
    MaxPlayers  int    `json:"max_players"`
    RoundTime   int    `json:"round_time"`
    MaxRounds   int    `json:"max_rounds"`
    PlayerCount int    `json:"player_count"`
}

// Add this new handler method
func (h *HTTPHandler) ValidateRoom(c *gin.Context) {
    var req JoinRoomRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    room, err := h.roomService.ValidateRoom(req.RoomCode)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get current player count
    playerCount := h.roomService.GetPlayerCount(req.RoomCode)

    response := JoinRoomResponse{
        RoomCode:    room.Code,
        MaxPlayers:  room.MaxPlayers,
        RoundTime:   room.RoundTime,
        MaxRounds:   room.MaxRounds,
        PlayerCount: playerCount,
    }

    c.JSON(http.StatusOK, response)
}

func (h *HTTPHandler) RegisterRoutes(r *gin.Engine) {
    // CORS middleware
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    api := r.Group("/api")
    {
        api.POST("/rooms", h.CreateRoom)
        api.GET("/rooms", h.GetActiveRooms)
        api.POST("/rooms/validate", h.ValidateRoom)
    }
}