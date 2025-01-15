// internal/repository/db.go

package repository

import (
	"fmt"

	"github.com/rohan03122001/quizzing/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
    *gorm.DB
}

func NewDatabase(host, port, user, password, dbname, sslmode string) (*Database, error) {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
        host, user, password, dbname, port, sslmode)

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Auto migrate the schemas
    err = db.AutoMigrate(
        &models.Room{},
        &models.Question{},
        &models.GameRound{},
        &models.PlayerAnswer{},
    )
    if err != nil {
        return nil, fmt.Errorf("failed to migrate database: %w", err)
    }

    return &Database{db}, nil
}