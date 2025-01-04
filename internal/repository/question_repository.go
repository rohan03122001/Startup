package repository

import "github.com/rohan03122001/quizzing/internal/models"

type QuestionRepository struct {
	db *Database
}

func NewQuestionRepository(db *Database) *QuestionRepository {
	return &QuestionRepository{
		db: db,
	}
}

func (r *QuestionRepository) Create(question models.Question) error{
	return r.db.Create(question).Error
}


func(r *QuestionRepository) GetRandom() (*models.Question, error){
	var question models.Question
	
	if err:= r.db.Order("RANDOM()").First(&question).Error; err!=nil{
		return nil, err
	}

	return &question, nil
}

// GetByCategory gets questions from a specific category
func (r *QuestionRepository) GetByCategory(category string) ([]models.Question, error) {
    var questions []models.Question
    if err := r.db.Where("category = ?", category).Find(&questions).Error; err != nil {
        return nil, err
    }
    return questions, nil
}

// GetByDifficulty gets questions of a specific difficulty
func (r *QuestionRepository) GetByDifficulty(difficulty int) ([]models.Question, error) {
    var questions []models.Question
    if err := r.db.Where("difficulty = ?", difficulty).Find(&questions).Error; err != nil {
        return nil, err
    }
    return questions, nil
}
