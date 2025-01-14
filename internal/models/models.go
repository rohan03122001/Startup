// internal/models/models.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a player
type User struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key"`
    Username  string    `gorm:"unique;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Room represents a game room
type Room struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key"`
    Code      string    `gorm:"unique;not null"`     // Room join code (e.g., "ABC123")
    Status    string    `gorm:"not null"`            // "waiting", "playing", "finished"
    CreatedAt time.Time
    EndedAt   *time.Time
    CreatedBy uuid.UUID `gorm:"type:uuid;not null"`  // User who created the room

    // Game Settings
    MaxPlayers   int `gorm:"default:10"`   // Maximum players allowed
    RoundTime    int `gorm:"default:30"`   // Seconds per round
    MaxRounds    int `gorm:"default:5"`    // Number of rounds
    CurrentRound int `gorm:"default:0"`    // Current round number
}

// Question represents a quiz question
type Question struct {
    ID         uuid.UUID `gorm:"type:uuid;primary_key"`
    Content    string    `gorm:"not null"`           // Question text
    Answer     string    `gorm:"not null"`           // Correct answer
    Category   string                                // Question category
    Difficulty int      `gorm:"default:1"`          // 1-5 difficulty scale
    CreatedAt  time.Time
}

// GameRound represents a single round in a game
type GameRound struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key"`
    RoomID      uuid.UUID `gorm:"type:uuid;not null"`
    QuestionID  uuid.UUID `gorm:"type:uuid;not null"`
    StartTime   time.Time
    EndTime     time.Time
    RoundNumber int    `gorm:"not null"`
    State       string `gorm:"not null;default:'waiting'"` // "waiting", "active", "finished"
    
    // Tracks how many correct answers received
    AnswerCount int `gorm:"default:0"`
}

// PlayerAnswer represents a player's answer in a round
type PlayerAnswer struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key"`
    RoundID     uuid.UUID `gorm:"type:uuid;not null"`
    UserID      uuid.UUID `gorm:"type:uuid;not null"`
    Answer      string    `gorm:"not null"`
    Score       int       `gorm:"default:0"`
    AnswerOrder int       `gorm:"not null"`    // Order in which answer was received
    AnsweredAt  time.Time
}

// Add these hooks to generate UUIDs for new records
func (u *User) BeforeCreate() error {
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    return nil
}

func (r *Room) BeforeCreate() error {
    if r.ID == uuid.Nil {
        r.ID = uuid.New()
    }
    return nil
}

func (q *Question) BeforeCreate() error {
    if q.ID == uuid.Nil {
        q.ID = uuid.New()
    }
    return nil
}

func (gr *GameRound) BeforeCreate() error {
    if gr.ID == uuid.Nil {
        gr.ID = uuid.New()
    }
    return nil
}

func (pa *PlayerAnswer) BeforeCreate() error {
    if pa.ID == uuid.Nil {
        pa.ID = uuid.New()
    }
    return nil
}