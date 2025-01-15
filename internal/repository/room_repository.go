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
    return &RoomRepository{db: db}
}

// CreateRoom creates a new room
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

// UpdateStatus updates a room's status
func (r *RoomRepository) UpdateStatus(roomID string, status string) error {
    log.Printf("Updating room %s status to: %s", roomID, status)
    return r.db.Model(&models.Room{}).
        Where("id = ?", roomID).
        Update("status", status).Error
}

// UpdateCurrentRound increments the current round number
func (r *RoomRepository) UpdateCurrentRound(roomID string) error {
    log.Printf("Incrementing current round for room: %s", roomID)
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

func (r *RoomRepository) GetByID(id string) (*models.Room, error) {
    log.Printf("Fetching room by ID: %s", id)
    var room models.Room
    err := r.db.Where("id = ?", id).First(&room).Error
    if err != nil {
        log.Printf("Error fetching room %s: %v", id, err)
        return nil, err
    }
    return &room, nil
}