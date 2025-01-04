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


//Get Active Rooms
func(r *RoomRepository) GetActive() ([]models.Room, error){
	var rooms []models.Room

	if err:=r.db.Where("status = ?","waiting").Or("status = ?", "playing").Find(&rooms).Error; err!=nil{
		return nil, err
	}

	return rooms, nil
}

//update Room Status
func(r *RoomRepository) UpdateStatus(RoomID string, status string) error{
	return r.db.Model(&models.Room{}).Where("id=?",RoomID).Update("status", status).Error
}


// Delete removes a room
func (r *RoomRepository) Delete(id string) error {
    return r.db.Delete(&models.Room{}, "id = ?", id).Error
}