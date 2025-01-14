package repository

import "github.com/rohan03122001/quizzing/internal/models"

type GameRoundRepository struct {
	db *Database
}

func NewGameRoundRepository(db *Database) *GameRoundRepository {
	return &GameRoundRepository{db: db}
}

func (r *GameRoundRepository) CreateRound(round *models.GameRound)error{
	return r.db.Create(round).Error
}

func(r *GameRoundRepository) GetCurrentRound(roomID string)(*models.GameRound, error){
	var round models.GameRound
	err := r.db.Where("room_id = ? AND state = ?", roomID, "active").First(&round).Error
	if err!=nil{
		return nil, err
	}
	return &round, nil
}


func(r *GameRoundRepository)SaveAnswer(answer *models.PlayerAnswer)error{
	return r.db.Create(answer).Error
}

func(r *GameRoundRepository)UpdateRound(round *models.GameRound) error{
	return r.db.Save(round).Error
}


func(r *GameRoundRepository)GetRoundAnswers(roomID string)([]models.PlayerAnswer, error){
	var answers []models.PlayerAnswer
	err:= r.db.Where("room_id = ?", roomID).Order("answer_order asc").Find(&answers).Error
	if err!= nil{
		return nil, err
	}
	return answers, nil
}

// UpdateAnswerCount increments the answer count for a round
func (r *GameRoundRepository) UpdateAnswerCount(roundID string) error {
    return r.db.Model(&models.GameRound{}).
        Where("id = ?", roundID).
        UpdateColumn("answer_count", r.db.Raw("answer_count + 1")).Error
}