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
    return &HTTPHandler{roomService: roomService}
}

func (h *HTTPHandler) CreateRoom(c *gin.Context) {
    room, err := h.roomService.CreateRoom()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, room)
}

func (h *HTTPHandler) GetRoom(c *gin.Context) {
    code := c.Param("code")
    room, err := h.roomService.GetRoom(code)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
        return
    }
    
    c.JSON(http.StatusOK, room)
}