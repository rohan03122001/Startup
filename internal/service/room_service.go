// internal/service/room_service.go

package service

import (
	"errors"
	"log"
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
        hub:      hub,
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
        existing, _ := s.roomRepo.GetByCode(roomCode)
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

    log.Printf("Creating new room with code: %s", roomCode)
    if err := s.roomRepo.CreateRoom(room); err != nil {
        return nil, err
    }

    return room, nil
}

// JoinRoom handles a player joining a room
// Update JoinRoom method
func (s *RoomService) JoinRoom(roomCode string, playerID string) (*models.Room, error) {
    log.Printf("Player %s trying to join room %s", playerID, roomCode)
    
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        log.Printf("Room not found: %s", roomCode)
        return nil, errors.New("room not found")
    }

    if room.Status != "waiting" {
        log.Printf("Room %s is not accepting players (status: %s)", roomCode, room.Status)
        return nil, errors.New("game already in progress")
    }

    // Check room capacity
    currentPlayers := s.hub.GetPlayerCount(room.ID.String())
    if currentPlayers >= room.MaxPlayers {
        log.Printf("Room %s is full (%d/%d players)", roomCode, currentPlayers, room.MaxPlayers)
        return nil, errors.New("room is full")
    }

    log.Printf("Player %s successfully joined room %s", playerID, roomCode)
    return room, nil
}

// Add method to get room by ID (needed for game handler)
func (s *RoomService) GetRoomByID(roomID string) (*models.Room, error) {
    return s.roomRepo.GetByID(roomID)
}

// StartGame initiates the game for a room
func (s *RoomService) StartGame(roomCode string) error {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        return errors.New("room not found")
    }

    if room.Status != "waiting" {
        return errors.New("game already started")
    }

    currentPlayers := s.hub.GetPlayerCount(room.ID.String())
    if currentPlayers < 2 {
        return errors.New("need at least 2 players to start")
    }

    log.Printf("Starting game in room %s with %d players", roomCode, currentPlayers)
    return s.roomRepo.UpdateStatus(room.ID.String(), "playing")
}

// GetActiveRooms returns all rooms waiting for players
func (s *RoomService) GetActiveRooms() ([]models.Room, error) {
    log.Println("Fetching active rooms")
    return s.roomRepo.GetActive()
}