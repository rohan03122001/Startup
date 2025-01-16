// internal/repository/game_round_repository.go

package repository

import (
	"log"

	"github.com/rohan03122001/quizzing/internal/models"
)

type GameRoundRepository struct {
    db *Database
}

func NewGameRoundRepository(db *Database) *GameRoundRepository {
    return &GameRoundRepository{
        db: db,
    }
}

// CreateRound starts a new round
func (r *GameRoundRepository) CreateRound(round *models.GameRound) error {
    log.Printf("Creating round %d for room %s", round.RoundNumber, round.RoomID)
    return r.db.Create(round).Error
}

// GetCurrentRound gets the active round for a room
func (r *GameRoundRepository) GetCurrentRound(roomID string) (*models.GameRound, error) {
    log.Printf("Fetching current round for room %s", roomID)
    var round models.GameRound
    err := r.db.Where("room_id = ? AND state = ?", roomID, "active").
        First(&round).Error
    if err != nil {
        log.Printf("Error fetching current round: %v", err)
        return nil, err
    }
    return &round, nil
}

// SaveAnswer records a player's answer
func (r *GameRoundRepository) SaveAnswer(answer *models.PlayerAnswer) error {
    log.Printf("Saving answer from player %s for round %s", answer.PlayerID, answer.RoundID)
    return r.db.Create(answer).Error
}

// GetRoundAnswers gets all answers for a round
func (r *GameRoundRepository) GetRoundAnswers(roundID string) ([]models.PlayerAnswer, error) {
    log.Printf("Fetching answers for round %s", roundID)
    var answers []models.PlayerAnswer
    err := r.db.Where("round_id = ?", roundID).
        Order("answer_order asc").
        Find(&answers).Error
    return answers, err
}

// UpdateAnswerCount increments the answer count
func (r *GameRoundRepository) UpdateAnswerCount(roundID string) error {
    log.Printf("Incrementing answer count for round %s", roundID)
    return r.db.Model(&models.GameRound{}).
        Where("id = ?", roundID).
        UpdateColumn("answer_count", r.db.Raw("answer_count + 1")).Error
}

// UpdateRoundState updates the state of a round
func (r *GameRoundRepository) UpdateRoundState(roundID string, state string) error {
    log.Printf("Updating round %s state to %s", roundID, state)
    return r.db.Model(&models.GameRound{}).
        Where("id = ?", roundID).
        Update("state", state).Error
}

// GetRoundScores gets scores for all players in a round
func (r *GameRoundRepository) GetRoundScores(roundID string) (map[string]int, error) {
    var answers []models.PlayerAnswer
    err := r.db.Where("round_id = ?", roundID).Find(&answers).Error
    if err != nil {
        return nil, err
    }

    scores := make(map[string]int)
    for _, answer := range answers {
        scores[answer.PlayerID] = answer.Score
    }
    return scores, nil
}

// GetRoomRounds gets all rounds for a room
func (r *GameRoundRepository) GetRoomRounds(roomID string) ([]models.GameRound, error) {
    var rounds []models.GameRound
    err := r.db.Where("room_id = ?", roomID).
        Order("round_number asc").
        Find(&rounds).Error
    if err != nil {
        log.Printf("Error fetching rounds for room %s: %v", roomID, err)
        return nil, err
    }
    return rounds, nil
}