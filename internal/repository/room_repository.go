// internal/repository/room_repository.go

package repository

import (
	"log"

	"github.com/rohan03122001/quizzing/internal/models"
)

type RoomRepository struct {
    db *Database
}

func NewRoomRepository(db *Database) *RoomRepository {
    return &RoomRepository{
        db: db,
    }
}

// CreateRoom adds a new room
func (r *RoomRepository) CreateRoom(room *models.Room) error {
    log.Printf("Creating room with code: %s", room.Code)
    return r.db.Create(room).Error
}

// GetByCode finds a room by its code
func (r *RoomRepository) GetByCode(code string) (*models.Room, error) {
    log.Printf("Fetching room with code: %s", code)
    var room models.Room
    err := r.db.Where("code = ?", code).First(&room).Error
    if err != nil {
        log.Printf("Error fetching room %s: %v", code, err)
        return nil, err
    }
    return &room, nil
}

// UpdateStatus updates room's status
func (r *RoomRepository) UpdateStatus(roomID string, status string) error {
    log.Printf("Updating status to %s for room %s", status, roomID)
    return r.db.Model(&models.Room{}).
        Where("id = ?", roomID).
        Update("status", status).Error
}

// UpdateCurrentRound increments the current round
func (r *RoomRepository) UpdateCurrentRound(roomID string) error {
    log.Printf("Incrementing round for room %s", roomID)
    return r.db.Model(&models.Room{}).
        Where("id = ?", roomID).
        UpdateColumn("current_round", r.db.Raw("current_round + 1")).Error
}

// GetActive gets all rooms that are waiting or playing
func (r *RoomRepository) GetActive() ([]models.Room, error) {
    log.Println("Fetching active rooms")
    var rooms []models.Room
    err := r.db.Where("status IN (?)", []string{"waiting", "playing"}).Find(&rooms).Error
    return rooms, err
}

// EndGame marks a room as finished
func (r *RoomRepository) EndGame(roomID string) error {
    log.Printf("Ending game for room %s", roomID)
    return r.db.Model(&models.Room{}).
        Where("id = ?", roomID).
        Updates(map[string]interface{}{
            "status": "finished",
            "ended_at": r.db.Raw("NOW()"),
        }).Error
}

// UpdateRoom updates room settings
func (r *RoomRepository) UpdateRoom(room *models.Room) error {
    log.Printf("Updating room %s settings", room.Code)
    return r.db.Save(room).Error
}