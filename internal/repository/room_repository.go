package repository

import "github.com/rohan03122001/quizzing/internal/models"

type RoomRepository struct {
	db *Database
}

func NewRoomRepository(db *Database) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) CreateRoom(room *models.Room) error{
	return r.db.Create(room).Error
}

func(r *RoomRepository) GetRoomByCode(code int) (*models.Room, error){
	var room models.Room

	if err:= r.db.First(&room,"code=?",code).Error; err!=nil{
		return nil, err
	}

	return &room, nil
}

