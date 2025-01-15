package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
    ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
    Code         string    `json:"code" gorm:"unique;not null"`
    Status       string    `json:"status"`
    MaxPlayers   int       `json:"max_players"`
    RoundTime    int       `json:"round_time"`
    MaxRounds    int       `json:"max_rounds"`
    CurrentRound int       `json:"current_round"`
    CreatedAt    time.Time `json:"created_at"`
    EndedAt      *time.Time `json:"ended_at,omitempty"`
}