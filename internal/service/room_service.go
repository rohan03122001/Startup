// internal/service/room_service.go

package service

import (
	"errors"
	"log"
	"math/rand"
	"time"

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
        Code:         roomCode,
        Status:       "waiting",
        MaxPlayers:   10,    // Default settings
        RoundTime:    30,    // 30 seconds per round
        MaxRounds:    5,     // 5 rounds per game
        LastActivity: time.Now(), // Explicitly set last activity time
    }

    if err := s.roomRepo.CreateRoom(room); err != nil {
        log.Printf("Failed to create room: %v", err)
        return nil, errors.New("failed to create room")
    }

    log.Printf("Created room with code: %s", roomCode)
    return room, nil
}

// JoinRoom updated to check active players only
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

    // Check room capacity using active player count
    activePlayerCount := s.hub.GetActivePlayerCount(room.Code)
    if activePlayerCount >= room.MaxPlayers {
        log.Printf("Room %s is full (%d/%d active players)", 
            roomCode, activePlayerCount, room.MaxPlayers)
        return nil, errors.New("room is full")
    }

    // Check if player is trying to rejoin (disconnected state)
    existingClient := s.hub.FindClientByID(room.Code, playerID)
    if existingClient != nil && existingClient.Status == "disconnected" {
        log.Printf("Player %s is rejoining room %s after disconnection", playerID, roomCode)
        // The reconnection will be handled by the websocket handler
    }

    // Update last activity
    if err := s.roomRepo.UpdateLastActivity(room.ID.String()); err != nil {
        log.Printf("Error updating room activity: %v", err)
    }

    log.Printf("Player %s successfully joined room %s", playerID, roomCode)
    return room, nil
}

// StartGame updated to check active players only
func (s *RoomService) StartGame(roomCode string) error {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        return errors.New("room not found")
    }

    if room.Status != "waiting" {
        return errors.New("game already started")
    }

    activePlayerCount := s.hub.GetActivePlayerCount(room.Code)
    log.Printf("Current active players in room %s: %d", roomCode, activePlayerCount)
    
    if activePlayerCount < 2 {
        return errors.New("need at least 2 active players to start")
    }

    err = s.roomRepo.UpdateStatus(room.ID.String(), "playing")
    if err != nil {
        return errors.New("failed to start game")
    }

    // Update last activity
    if err := s.roomRepo.UpdateLastActivity(room.ID.String()); err != nil {
        log.Printf("Error updating room activity: %v", err)
    }

    log.Printf("Started game in room %s with %d active players", roomCode, activePlayerCount)
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

func (s *RoomService) ValidateRoom(roomCode string) (*models.Room, error) {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        return nil, errors.New("room not found")
    }

    if room.Status != "waiting" {
        return nil, errors.New("game already in progress")
    }

    // Check active players capacity 
    activePlayerCount := s.hub.GetActivePlayerCount(room.Code)
    if activePlayerCount >= room.MaxPlayers {
        return nil, errors.New("room is full")
    }

    return room, nil
}

func (s *RoomService) GetPlayerCount(roomCode string) int {
    return s.hub.GetActivePlayerCount(roomCode)
}

// UpdateRoomActivity updates the last activity timestamp for a room
func (s *RoomService) UpdateRoomActivity(roomCode string) error {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        log.Printf("Error finding room %s to update activity: %v", roomCode, err)
        return err
    }
    
    if err := s.roomRepo.UpdateLastActivity(room.ID.String()); err != nil {
        log.Printf("Error updating room %s activity: %v", roomCode, err)
        return err
    }
    
    return nil
}