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

    c.JSON(http.StatusCreated, gin.H{
        "room_code": room.Code,
        "max_players": room.MaxPlayers,
        "round_time": room.RoundTime,
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
    }
}