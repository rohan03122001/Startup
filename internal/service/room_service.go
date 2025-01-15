package service

import (
	"math/rand"

	"github.com/rohan03122001/quizzing/internal/models"
	"github.com/rohan03122001/quizzing/internal/repository"
)

type RoomService struct {
    roomRepo *repository.RoomRepository
}

func NewRoomService(roomRepo *repository.RoomRepository) *RoomService {
    return &RoomService{roomRepo: roomRepo}
}

func (s *RoomService) CreateRoom() (*models.Room, error) {
    room := &models.Room{
        Code:       generateRoomCode(),
        Status:     "waiting",
        MaxPlayers: 10,
        RoundTime:  30,
        MaxRounds:  5,
    }
    
    if err := s.roomRepo.Create(room); err != nil {
        return nil, err
    }
    
    return room, nil
}

func (s *RoomService) GetRoom(code string) (*models.Room, error) {
    return s.roomRepo.FindByCode(code)
}

func generateRoomCode() string {
    const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    code := make([]byte, 6)
    for i := range code {
        code[i] = letters[rand.Intn(len(letters))]
    }
    return string(code)
}
