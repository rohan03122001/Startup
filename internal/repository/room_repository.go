// internal/repository/room_repository.go

package repository

import (
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
    return r.db.Create(room).Error
}

// GetRoomByCode finds a room by its code
func (r *RoomRepository) GetRoomByCode(code string) (*models.Room, error) {
    var room models.Room
    err := r.db.Where("code = ?", code).First(&room).Error
    if err != nil {
        return nil, err
    }
    return &room, nil
}

// UpdateStatus updates a room's status
func (r *RoomRepository) UpdateStatus(roomID string, status string) error {
    return r.db.Model(&models.Room{}).Where("id = ?", roomID).Update("status", status).Error
}

// GetActive gets all rooms that are waiting or playing
func (r *RoomRepository) GetActive() ([]models.Room, error) {
    var rooms []models.Room
    err := r.db.Where("status IN (?)", []string{"waiting", "playing"}).Find(&rooms).Error
    return rooms, err
}