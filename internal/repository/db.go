// internal/repository/db.go

package repository

import (
	"fmt"
	"log"

	"github.com/rohan03122001/quizzing/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
    // Construct DSN (Data Source Name)
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
        config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode,
    )

    log.Printf("Connecting to database on %s:%s", config.Host, config.Port)

    // Configure GORM
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info), // Set logging level
    }

    // Open connection
    db, err := gorm.Open(postgres.Open(dsn), gormConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Get underlying SQL DB to set pool settings
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get database instance: %w", err)
    }

    // Set connection pool settings
    sqlDB.SetMaxIdleConns(10)  // Maximum number of idle connections
    sqlDB.SetMaxOpenConns(100) // Maximum number of open connections

    log.Println("Running database migrations...")
    
    // Auto migrate schemas
    err = db.AutoMigrate(
        &models.Room{},
        &models.Question{},
        &models.GameRound{},
        &models.PlayerAnswer{},
    )
    if err != nil {
        return nil, fmt.Errorf("failed to migrate database: %w", err)
    }

    log.Println("Database migrations completed successfully")

    return &Database{db}, nil
}

// GetDB returns the underlying GORM DB instance
func (d *Database) GetDB() *gorm.DB {
    return d.DB
}

// Close closes the database connection
func (d *Database) Close() error {
    sqlDB, err := d.DB.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}