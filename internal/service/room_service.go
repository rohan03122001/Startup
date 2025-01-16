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
    hub      *websocket.Hub  // For real-time updates
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
func (s *RoomService) CreateRoom() (*models.Room, error) {
    // Generate unique room code
    var roomCode string
    for i := 0; i < 5; i++ { // Try 5 times to generate unique code
        roomCode = s.generateRoomCode()
        existing, _ := s.roomRepo.GetByCode(roomCode)
        if existing == nil {
            break
        }
    }

    // Create new room
    room := &models.Room{
        Code:       roomCode,
        Status:     "waiting",
        MaxPlayers: 10,    // Default settings
        RoundTime:  30,    // 30 seconds per round
        MaxRounds:  5,     // 5 rounds per game
    }

    if err := s.roomRepo.CreateRoom(room); err != nil {
        log.Printf("Failed to create room: %v", err)
        return nil, errors.New("failed to create room")
    }

    log.Printf("Created room with code: %s", roomCode)
    return room, nil
}

// internal/service/room_service.go

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
    currentPlayers := s.hub.GetPlayerCount(room.Code)  // Changed from ID to Code
    if currentPlayers >= room.MaxPlayers {
        log.Printf("Room %s is full (%d/%d players)", roomCode, currentPlayers, room.MaxPlayers)
        return nil, errors.New("room is full")
    }

    // Update last activity
    if err := s.roomRepo.UpdateLastActivity(room.ID.String()); err != nil {
        log.Printf("Error updating room activity: %v", err)
    }

    log.Printf("Player %s successfully joined room %s", playerID, roomCode)
    return room, nil
}

// StartGame updated to use room code
func (s *RoomService) StartGame(roomCode string) error {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        return errors.New("room not found")
    }

    if room.Status != "waiting" {
        return errors.New("game already started")
    }

    currentPlayers := s.hub.GetPlayerCount(room.Code)  // Changed from ID to Code
    log.Printf("Current players in room %s: %d", roomCode, currentPlayers)
    
    if currentPlayers < 2 {
        return errors.New("need at least 2 players to start")
    }

    err = s.roomRepo.UpdateStatus(room.ID.String(), "playing")
    if err != nil {
        return errors.New("failed to start game")
    }

    // Update last activity
    if err := s.roomRepo.UpdateLastActivity(room.ID.String()); err != nil {
        log.Printf("Error updating room activity: %v", err)
    }

    log.Printf("Started game in room %s with %d players", roomCode, currentPlayers)
    return nil
}

// GetActiveRooms returns all rooms waiting for players
func (s *RoomService) GetActiveRooms() ([]models.Room, error) {
    log.Println("Fetching active rooms")
    return s.roomRepo.GetActive()
}

// EndGame finishes a game
func (s *RoomService) EndGame(roomCode string) error {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        return errors.New("room not found")
    }

    log.Printf("Ending game in room %s", roomCode)
    return s.roomRepo.EndGame(room.ID.String())
}

// GetRoom retrieves a room by its code
func (s *RoomService) GetRoom(roomCode string) (*models.Room, error) {
    return s.roomRepo.GetByCode(roomCode)
}