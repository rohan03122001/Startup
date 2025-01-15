// internal/repository/game_round_repository.go

package repository

import "github.com/rohan03122001/quizzing/internal/models"

type GameRoundRepository struct {
    db *Database
}

func NewGameRoundRepository(db *Database) *GameRoundRepository {
    return &GameRoundRepository{db: db}
}

// CreateRound creates a new game round
func (r *GameRoundRepository) CreateRound(round *models.GameRound) error {
    return r.db.Create(round).Error
}

// GetCurrentRound gets the active round for a room
func (r *GameRoundRepository) GetCurrentRound(roomID string) (*models.GameRound, error) {
    var round models.GameRound
    err := r.db.Where("room_id = ? AND state = ?", roomID, "active").First(&round).Error
    if err != nil {
        return nil, err
    }
    return &round, nil
}

// SaveAnswer saves a player's answer
func (r *GameRoundRepository) SaveAnswer(answer *models.PlayerAnswer) error {
    return r.db.Create(answer).Error
}

// GetRoundAnswers gets all answers for a round
func (r *GameRoundRepository) GetRoundAnswers(roundID string) ([]models.PlayerAnswer, error) {
    var answers []models.PlayerAnswer
    err := r.db.Where("round_id = ?", roundID).Order("answer_order asc").Find(&answers).Error
    return answers, err
}

// UpdateAnswerCount increments answer count for a round
func (r *GameRoundRepository) UpdateAnswerCount(roundID string) error {
    return r.db.Model(&models.GameRound{}).
        Where("id = ?", roundID).
        UpdateColumn("answer_count", r.db.Raw("answer_count + 1")).Error
}