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

type CreateRoomResponse struct {
    RoomCode   string `json:"room_code"`
    MaxPlayers int    `json:"max_players"`
    RoundTime  int    `json:"round_time"`
}

// CreateRoom creates a new game room
func (h *HTTPHandler) CreateRoom(c *gin.Context) {
    room, err := h.roomService.CreateRoom(10) // Default 10 players
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, CreateRoomResponse{
        RoomCode:   room.Code,
        MaxPlayers: room.MaxPlayers,
        RoundTime:  room.RoundTime,
    })
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

// RegisterRoutes sets up HTTP routes
func (h *HTTPHandler) RegisterRoutes(r *gin.Engine) {
    api := r.Group("/api")
    {
        api.POST("/rooms", h.CreateRoom)
        api.GET("/rooms", h.GetActiveRooms)
    }
}