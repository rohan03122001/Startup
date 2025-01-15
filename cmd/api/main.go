package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/rohan03122001/quizzing/internal/config"
	"github.com/rohan03122001/quizzing/internal/handlers"
	"github.com/rohan03122001/quizzing/internal/repository"
	"github.com/rohan03122001/quizzing/internal/service"
)

func main() {
    // Initialize configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize database
    db, err := repository.NewDB(*cfg.Database)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Initialize repositories
    roomRepo := repository.NewRoomRepository(db)

    // Initialize services
    roomService := service.NewRoomService(roomRepo)

    // Initialize handlers
    roomHandler := handlers.NewHTTPHandler(roomService)

    // Setup router
    r := gin.Default()
    
    // Register routes
    api := r.Group("/api")
    {
        api.POST("/rooms", roomHandler.CreateRoom)
        api.GET("/rooms/:code", roomHandler.GetRoom)
    }

    // Start server
    if err := r.Run(cfg.Server.Address); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}