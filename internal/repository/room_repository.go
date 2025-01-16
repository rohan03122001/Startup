// internal/repository/room_repository.go

package repository

import (
	"log"
	"time"

	"github.com/rohan03122001/quizzing/internal/models"
	"gorm.io/gorm"
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

// GetInactiveRooms gets rooms that haven't had activity recently
func (r *RoomRepository) GetInactiveRooms(before time.Time) ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.Where("last_activity < ? AND status IN (?)", 
		before, []string{"waiting", "playing"}).
		Find(&rooms).Error
	return rooms, err
}

// DeleteRoom removes a room and its related data
func (r *RoomRepository) DeleteRoom(roomID string) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // First get all rounds for this room
        var rounds []models.GameRound
        if err := tx.Where("room_id = ?", roomID).Find(&rounds).Error; err != nil {
            return err
        }

        // Get round IDs
        roundIDs := make([]string, len(rounds))
        for i, round := range rounds {
            roundIDs[i] = round.ID.String()
        }

        // Delete player answers for these rounds
        if len(roundIDs) > 0 {
            if err := tx.Where("round_id IN ?", roundIDs).Delete(&models.PlayerAnswer{}).Error; err != nil {
                return err
            }
        }

        // Delete game rounds
        if err := tx.Where("room_id = ?", roomID).Delete(&models.GameRound{}).Error; err != nil {
            return err
        }

        // Finally delete the room
        return tx.Where("id = ?", roomID).Delete(&models.Room{}).Error
    })
}

// UpdateLastActivity updates the room's last activity timestamp
func (r *RoomRepository) UpdateLastActivity(roomID string) error {
	return r.db.Model(&models.Room{}).
		Where("id = ?", roomID).
		Update("last_activity", time.Now()).Error
}