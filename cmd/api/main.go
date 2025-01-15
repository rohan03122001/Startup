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
	"github.com/rohan03122001/quizzing/internal/config"
	"github.com/rohan03122001/quizzing/internal/handlers"
	"github.com/rohan03122001/quizzing/internal/repository"
	"github.com/rohan03122001/quizzing/internal/service"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize database
    db, err := repository.NewDatabase(
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.DBName,
        cfg.Database.SSLMode,
    )
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Initialize repositories
    roomRepo := repository.NewRoomRepository(db)
    questionRepo := repository.NewQuestionRepository(db)
    roundRepo := repository.NewGameRoundRepository(db)

    // Initialize WebSocket hub
    hub := websocket.NewHub()
    go hub.Run()

    // Initialize services
    roomService := service.NewRoomService(roomRepo, hub)
    gameService := service.NewGameService(roomRepo, questionRepo, roundRepo, hub)

    // Initialize handlers
    gameHandler := handlers.NewGameHandler(gameService, roomService, hub)
    httpHandler := handlers.NewHTTPHandler(roomService)
    wsHandler := websocket.NewHandler(hub)

    // Set up router
    gin.SetMode(cfg.Server.Mode)
    router := gin.Default()

    // CORS middleware
    router.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    // Register routes
    httpHandler.RegisterRoutes(router)
    
    // WebSocket route
    router.GET("/ws", wsHandler.HandleConnection)

    // Create server
    srv := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: router,
    }

    // Start server in goroutine
    go func() {
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