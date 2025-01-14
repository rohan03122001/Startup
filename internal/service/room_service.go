package service

import (
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/rohan03122001/quizzing/internal/models"
	"github.com/rohan03122001/quizzing/internal/repository"
	"github.com/rohan03122001/quizzing/internal/websockets"
)

type RoomService struct {
	roomRepo *repository.RoomRepository
	hub *websockets.Hub
}


func NewRoomService(roomRepo *repository.RoomRepository, hub *websockets.Hub) *RoomService{
	return &RoomService{
		roomRepo: roomRepo,
		hub: hub,
	}
}

func(s *RoomService)GenerateRoomCode() string{
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"	//doesnt have same looking chars
	code := make([]byte, 6)
	for i := range code{
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func(s *RoomService)CreateRoom(creatorID uuid.UUID, maxPlayer int)(*models.Room, error){
	var roomCode string
	for i:= 0; i< 5 ; i++ {
		roomCode =s.GenerateRoomCode()
		existing, _ := s.roomRepo.GetRoomByCode(roomCode) 
		if(existing == nil){
			break
		}
	}
	room := &models.Room{
		Code: roomCode,
		Status: "waiting",
		CreatedBy: creatorID,
		MaxPlayers: maxPlayer,
		RoundTime: 30,
		MaxRounds: 5,
	}
	if err:= s.roomRepo.CreateRoom(room); err!=nil{
		return nil, err
	}

	return room, nil

}

func (s *RoomService) JoinRoom(roomCode string, userID string) (*models.Room, error) {
    room, err := s.roomRepo.GetRoomByCode(roomCode)
    if err != nil {
        return nil, errors.New("room not found")
    }

    if room.Status != "waiting" {
        return nil, errors.New("game already in progress")
    }

    // Check room capacity
    currentPlayers := s.hub.GetPlayerCount(room.ID.String())
    if currentPlayers >= room.MaxPlayers {
        return nil, errors.New("room is full")
    }

    return room, nil
}

func(s *RoomService)JoinGame(roomCode string)(*models.Room, error){

	room, err := s.roomRepo.GetRoomByCode(roomCode)
	if err!= nil{
		return nil,errors.New("room not found")
	}

	if room.Status != "waiting"{
		return nil, errors.New("game in progress")
	}

	//TODO

	currentPlayers:= s.hub.GetPlayerCount(room.ID.String())
	if currentPlayers >= room.MaxPlayers{
		return nil, errors.New("room is full")
	}

	return room, nil
}

func(s *RoomService)StartGame(roomCode string)error{
	room, err := s.roomRepo.GetRoomByCode(roomCode)
	if err!= nil{
		return errors.New("room not found")
	}

	if room.Status != "waiting" {
        return errors.New("game already started")
    }

	currentPlayers:= s.hub.GetPlayerCount(room.ID.String())
	if currentPlayers < 2{
		return errors.New("need at least 2 players to start")
	}
	return s.roomRepo.UpdateStatus(room.ID.String(),"playing")
}


func(s *RoomService)GetAllActiveRooms()([]models.Room, error){
	return s.roomRepo.GetActive()
}


func(s *RoomService)EndGame(roomCode string)error{
	room, err := s.roomRepo.GetRoomByCode(roomCode)
	if err != nil {
        return errors.New("room not found")
    }

	endTime:= time.Now()
	room.EndedAt  = &endTime
	room.Status = "finished"

	return s.roomRepo.UpdateStatus(room.ID.String(), "finished")
}