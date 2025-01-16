// internal/models/models.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Room represents a game room
type Room struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key"`
    Code         string    `gorm:"unique;not null"`      // 6-character room code (e.g., "ABC123")
    Status       string    `gorm:"not null"`            // "waiting", "playing", "finished"
    MaxPlayers   int       `gorm:"default:10"`          // Maximum players allowed
    RoundTime    int       `gorm:"default:30"`          // Seconds per round
    MaxRounds    int       `gorm:"default:2"`           // Number of rounds
    CurrentRound int       `gorm:"default:0"`           // Current round number
    CreatedAt    time.Time
    EndedAt      *time.Time
}

// Question represents a quiz question
type Question struct {
    ID         uuid.UUID `gorm:"type:uuid;primary_key"`
    Content    string    `gorm:"not null"`           // Question text
    Answer     string    `gorm:"not null"`           // Correct answer
    CreatedAt  time.Time
}

// GameRound represents a single round in a game
type GameRound struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key"`
    RoomID       uuid.UUID `gorm:"type:uuid;not null"`
    QuestionID   uuid.UUID `gorm:"type:uuid;not null"`
    StartTime    time.Time
    EndTime      time.Time
    RoundNumber  int     `gorm:"not null"`
    State        string  `gorm:"not null;default:'waiting'"` // "waiting", "active", "finished"
    AnswerCount  int     `gorm:"default:0"`                 // Number of answers received
}

// PlayerAnswer represents a player's answer in a round
type PlayerAnswer struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key"`
    RoundID     uuid.UUID `gorm:"type:uuid;not null"`
    PlayerID    string    `gorm:"not null"`           // Client ID from WebSocket
    Answer      string    `gorm:"not null"`
    Score       int       `gorm:"default:0"`
    AnswerOrder int       `gorm:"not null"`           // Order in which answer was received
    AnsweredAt  time.Time
}

// GameSettings represents game settings
type GameSettings struct {
    MaxRounds  int `json:"max_rounds"`
    RoundTime  int `json:"round_time"`
}

// BeforeCreate hooks to generate UUIDs
func (r *Room) BeforeCreate(tx *gorm.DB) error {
    if r.ID == uuid.Nil {
        r.ID = uuid.New()
    }
    return nil
}

func (q *Question) BeforeCreate(tx *gorm.DB) error {
    if q.ID == uuid.Nil {
        q.ID = uuid.New()
    }
    return nil
}

func (gr *GameRound) BeforeCreate(tx *gorm.DB) error {
    if gr.ID == uuid.Nil {
        gr.ID = uuid.New()
    }
    return nil
}

func (pa *PlayerAnswer) BeforeCreate(tx *gorm.DB) error {
    if pa.ID == uuid.Nil {
        pa.ID = uuid.New()
    }
    return nil
}