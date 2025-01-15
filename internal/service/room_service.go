// internal/service/room_service.go

package service

import (
	"errors"
	"math/rand"

	"github.com/rohan03122001/quizzing/internal/models"
	"github.com/rohan03122001/quizzing/internal/repository"
	"github.com/rohan03122001/quizzing/internal/websocket"
)

type RoomService struct {
    roomRepo *repository.RoomRepository
    hub      *websocket.Hub
}

func NewRoomService(roomRepo *repository.RoomRepository, hub *websocket.Hub) *RoomService {
    return &RoomService{
        roomRepo: roomRepo,
        hub:     hub,
    }
}

// generateRoomCode creates a unique 6-character room code
func (s *RoomService) generateRoomCode() string {
    const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed similar looking characters
    code := make([]byte, 6)
    for i := range code {
        code[i] = charset[rand.Intn(len(charset))]
    }
    return string(code)
}

// CreateRoom creates a new game room
func (s *RoomService) CreateRoom(maxPlayers int) (*models.Room, error) {
    // Generate unique room code
    var roomCode string
    for i := 0; i < 5; i++ { // Try 5 times
        roomCode = s.generateRoomCode()
        existing, _ := s.roomRepo.GetRoomByCode(roomCode)
        if existing == nil {
            break
        }
    }

    room := &models.Room{
        Code:       roomCode,
        Status:     "waiting",
        MaxPlayers: maxPlayers,
        RoundTime:  30,  // 30 seconds per round
        MaxRounds:  5,   // 5 rounds per game
    }

    if err := s.roomRepo.CreateRoom(room); err != nil {
        return nil, err
    }

    return room, nil
}

// JoinRoom handles a player joining a room
func (s *RoomService) JoinRoom(roomCode string) (*models.Room, error) {
    room, err := s.roomRepo.GetRoomByCode(roomCode)
    if err != nil {
        return nil, errors.New("room not found")
    }

    if room.Status != "waiting" {
        return nil, errors.New("game already in progress")
    }

    // Check if room is full
    currentPlayers := s.hub.GetPlayerCount(room.ID.String())
    if currentPlayers >= room.MaxPlayers {
        return nil, errors.New("room is full")
    }

    return room, nil
}

// StartGame initiates the game for a room
func (s *RoomService) StartGame(roomCode string) error {
    room, err := s.roomRepo.GetRoomByCode(roomCode)
    if err != nil {
        return errors.New("room not found")
    }

    if room.Status != "waiting" {
        return errors.New("game already started")
    }

    // Need at least 2 players to start
    currentPlayers := s.hub.GetPlayerCount(room.ID.String())
    if currentPlayers < 2 {
        return errors.New("need at least 2 players to start")
    }

    return s.roomRepo.UpdateStatus(room.ID.String(), "playing")
}

// GetActiveRooms returns all rooms waiting for players
func (s *RoomService) GetActiveRooms() ([]models.Room, error) {
    return s.roomRepo.GetActive()
}