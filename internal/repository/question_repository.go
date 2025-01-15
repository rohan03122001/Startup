// internal/repository/question_repository.go

package repository

import "github.com/rohan03122001/quizzing/internal/models"

type QuestionRepository struct {
    db *Database
}

func NewQuestionRepository(db *Database) *QuestionRepository {
    return &QuestionRepository{db: db}
}

// GetRandom gets a random question
func (r *QuestionRepository) GetRandom() (*models.Question, error) {
    var question models.Question
    err := r.db.Order("RANDOM()").First(&question).Error
    if err != nil {
        return nil, err
    }
    return &question, nil
}

// GetByID gets a specific question
func (r *QuestionRepository) GetByID(id string) (*models.Question, error) {
    var question models.Question
    err := r.db.First(&question, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &question, nil
}