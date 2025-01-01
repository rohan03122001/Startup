package models

import (
	"time"

	"github.com/google/uuid"
)

//user struct will represent a single user having ID and Username
type User struct{
	ID	uuid.UUID	`gorm:"type:UUID;primary_key;"`
	Username string `gorm:"unique;not null"`
	CreatedAt	time.Time
	UpdatedAt	time.Time
}

//we will create room

type Room struct{
	ID uuid.UUID	`gorm:"type:UUID;primary_key"`
	Code string	`gorm:"unique;not null"`
	Status string `gorm:"not null"`
	CreatedAt time.Time
	EndedAt	*time.Time
	CreatedBy	uuid.UUID	`gorm:"type:UUID;not null"`

	//game setting
	MaxPlayers int	`gorm:"default:10"`
	RoundTime	int	`gorm:"default:30"`
	MaxRound	int		`gorm:"default:5"`
	CurrentRound	int	`gorm:"default:0"`
}

type Question struct{
	ID uuid.UUID	`gorm:"type:UUID;primary_key"`
	Content string	`gorm:"not null"`
	Answer string	`gorm:"not null"`
	Category string
	Difficulty int	`gorm:"default:1"`
	CreatedAt time.Time
}


type GameRound struct{
	ID uuid.UUID	`gorm:"type:UUID;primary_key"`
	RoomID	uuid.UUID `gorm:"type:UUID;not null"`
	QuestionID uuid.UUID `gorm:"type:uuid;not null"`
	StartTime	time.Time
	EndTime	time.Time
	RoundNumber	int	`gorm:"not null"`
}

type PlayerAnswer struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
    RoundID   uuid.UUID `gorm:"type:uuid;not null"`
    UserID    uuid.UUID `gorm:"type:uuid;not null"`
    Answer    string    `gorm:"not null"`
    Score     int       `gorm:"default:0"`
    AnsweredAt time.Time
}


//initialise  creates a new instance with UUID

func(u *User)BeforeCreate() error {
	u.ID = uuid.New()
	return nil
}

func(r *Room)BeforeCreate() error { 
	r.ID = uuid.New()
	return nil
}

func(q *Question)BeforeCreate() error {
	q.ID = uuid.New()
	return nil
}