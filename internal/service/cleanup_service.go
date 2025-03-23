package service

import (
	"log"
	"time"

	"github.com/rohan03122001/quizzing/internal/repository"
	"github.com/rohan03122001/quizzing/internal/websocket"
)

const (
	inactiveRoomTimeout = 10 * time.Minute
	cleanupInterval    = 1 * time.Minute
)

type CleanupService struct {
	roomRepo *repository.RoomRepository
	hub      *websocket.Hub
}

func NewCleanupService(roomRepo *repository.RoomRepository, hub *websocket.Hub) *CleanupService {
	return &CleanupService{
		roomRepo: roomRepo,
		hub:      hub,
	}
}

func (s *CleanupService) StartCleanupRoutine() {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for range ticker.C {
			s.cleanupInactiveRooms()
		}
	}()
	log.Println("Room cleanup routine started")
}

func (s *CleanupService) cleanupInactiveRooms() {
	inactiveTime := time.Now().Add(-inactiveRoomTimeout)
	
	// Get rooms that need cleanup
	rooms, err := s.roomRepo.GetInactiveRooms(inactiveTime)
	if err != nil {
		log.Printf("Error fetching inactive rooms: %v", err)
		return
	}

	for _, room := range rooms {
		playerCount := s.hub.GetPlayerCount(room.Code)

		switch room.Status {
		case "waiting":
			// Delete rooms with no players that haven't had activity
			if playerCount == 0 {
				if err := s.roomRepo.DeleteRoom(room.ID.String()); err != nil {
					log.Printf("Error deleting inactive room %s: %v", room.Code, err)
					continue
				}
				log.Printf("Deleted inactive waiting room: %s", room.Code)
			}

		case "playing":
			// If no players in an active game, mark it as abandoned
			if playerCount == 0 {
				if err := s.roomRepo.UpdateStatus(room.ID.String(), "abandoned"); err != nil {
					log.Printf("Error marking room %s as abandoned: %v", room.Code, err)
					continue
				}
				log.Printf("Marked empty game room as abandoned: %s", room.Code)
			}
		}
	}
} 