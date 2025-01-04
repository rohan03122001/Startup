package repository

import (
	"fmt"

	"github.com/rohan03122001/quizzing/internal/config"
	"github.com/rohan03122001/quizzing/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct{
	*gorm.DB
}


func NewDatabase(cfg *config.Database)(*Database, error){
	dsn := fmt.Sprintf("host %s user%s password %s dbname %s port %s sslmode %s", cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil{
		return nil, fmt.Errorf("failed to connect database %w", err)
	}

	err = db.AutoMigrate(
		&models.User{},
        &models.Room{},
        &models.Question{},
        &models.GameRound{},
        &models.PlayerAnswer{},
    )
	
	if err != nil{
		return nil, fmt.Errorf("failed to migrate database %w", err)
	}

	return &Database{db}, nil
}