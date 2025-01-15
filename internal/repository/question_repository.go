// internal/repository/question_repository.go

package repository

import (
	"log"

	"github.com/rohan03122001/quizzing/internal/models"
)

type QuestionRepository struct {
    db *Database
}

func NewQuestionRepository(db *Database) *QuestionRepository {
    return &QuestionRepository{
        db: db,
    }
}

// CreateQuestion adds a new question
func (r *QuestionRepository) CreateQuestion(question *models.Question) error {
    log.Printf("Creating new question: %s", question.Content)
    return r.db.Create(question).Error
}

// GetRandom gets a random question
func (r *QuestionRepository) GetRandom() (*models.Question, error) {
    log.Println("Fetching random question")
    var question models.Question
    err := r.db.Order("RANDOM()").First(&question).Error
    if err != nil {
        log.Printf("Error fetching random question: %v", err)
        return nil, err
    }
    return &question, nil
}

// GetByID gets a specific question by ID
func (r *QuestionRepository) GetByID(id string) (*models.Question, error) {
    log.Printf("Fetching question by ID: %s", id)
    var question models.Question
    err := r.db.First(&question, "id = ?", id).Error
    if err != nil {
        log.Printf("Error fetching question %s: %v", id, err)
        return nil, err
    }
    return &question, nil
}

// GetQuestionCount returns total number of questions
func (r *QuestionRepository) GetQuestionCount() (int64, error) {
    var count int64
    err := r.db.Model(&models.Question{}).Count(&count).Error
    if err != nil {
        log.Printf("Error counting questions: %v", err)
        return 0, err
    }
    return count, nil
}