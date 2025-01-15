// internal/repository/db.go

package repository

import (
	"fmt"
	"log"

	"github.com/rohan03122001/quizzing/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
    *gorm.DB
}

type DBConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func NewDatabase(config *DBConfig) (*Database, error) {
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
        config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode,
    )

    log.Printf("Connecting to database on %s:%s", config.Host, config.Port)
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Auto migrate schemas
    log.Println("Running database migrations...")
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