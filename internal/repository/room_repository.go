package repository

import (
	"github.com/google/uuid"
	"github.com/rohan03122001/quizzing/internal/models"
	"gorm.io/gorm"
)

type RoomRepository struct {
    db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
    return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(room *models.Room) error {
    room.ID = uuid.New()
    return r.db.Create(room).Error
}

func (r *RoomRepository) FindByCode(code string) (*models.Room, error) {
    var room models.Room
    if err := r.db.Where("code = ?", code).First(&room).Error; err != nil {
        return nil, err
    }
    return &room, nil
}