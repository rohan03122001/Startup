package repository

import (
	"fmt"

	"github.com/rohan03122001/quizzing/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB(cfg config.Database) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
    
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}