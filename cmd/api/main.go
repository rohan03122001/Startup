// cmd/api/main.go

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohan03122001/quizzing/internal/handlers"
	"github.com/rohan03122001/quizzing/internal/repository"
	"github.com/rohan03122001/quizzing/internal/service"
	"github.com/rohan03122001/quizzing/internal/websocket"
)

func main() {
    // Set Gin mode
    gin.SetMode(gin.DebugMode)

    // Initialize database
    db, err := repository.NewDatabase(&repository.DBConfig{
        Host:     "localhost",
        Port:     "5434",  // Match the port from docker-compose
        User:     "postgres",
        Password: "postgres",
        DBName:   "quiz_app",
        SSLMode:  "disable",
    })
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // Initialize WebSocket hub
    hub := websocket.NewHub()
    go hub.Run()

    // Initialize repositories
    roomRepo := repository.NewRoomRepository(db)
    questionRepo := repository.NewQuestionRepository(db)
    roundRepo := repository.NewGameRoundRepository(db)

    // Initialize services
    roomService := service.NewRoomService(roomRepo, hub)
    gameService := service.NewGameService(roomRepo, questionRepo, roundRepo, hub)

    // Initialize handlers
    httpHandler := handlers.NewHTTPHandler(roomService)
    gameHandler := handlers.NewGameHandler(gameService, roomService, hub)
    wsHandler := handlers.NewWebSocketHandler(hub, gameHandler)

    // Setup Gin router
    router := gin.Default()

    // Register routes
    httpHandler.RegisterRoutes(router)
    wsHandler.RegisterRoutes(router)

    // Create server
    srv := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }

    // Start server in goroutine
    go func() {
        log.Printf("Server starting on port 8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    // Give 5 seconds for graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server exited")
}