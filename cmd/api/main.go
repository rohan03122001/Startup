// cmd/api/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohan03122001/quizzing/internal/config"
	"github.com/rohan03122001/quizzing/internal/handlers"
	"github.com/rohan03122001/quizzing/internal/repository"
	"github.com/rohan03122001/quizzing/internal/service"
	"github.com/rohan03122001/quizzing/internal/websocket"
)

func main() {
    // Set Gin mode based on environment
    ginMode := os.Getenv("GIN_MODE")
    if ginMode == "" {
        ginMode = gin.DebugMode
    }
    gin.SetMode(ginMode)

    // Get database configuration from environment variables
    dbConfig, err := config.GetDBConfig()
    if err != nil {
        log.Fatalf("Failed to get database config: %v", err)
    }

    // Initialize database
    db, err := repository.NewDatabase(&repository.DBConfig{
        Host:     dbConfig.Host,
        Port:     strconv.Itoa(dbConfig.Port),
        User:     dbConfig.User,
        Password: dbConfig.Password,
        DBName:   dbConfig.DBName,
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
    
    // Set room service on hub for activity updates
    hub.SetRoomService(roomService)
    
    gameService := service.NewGameService(roomRepo, questionRepo, roundRepo, hub)
    cleanupService := service.NewCleanupService(roomRepo, hub)
    cleanupService.StartCleanupRoutine()

    // Initialize handlers
    httpHandler := handlers.NewHTTPHandler(roomService)
    gameHandler := handlers.NewGameHandler(gameService, roomService, hub)
    wsHandler := handlers.NewWebSocketHandler(hub, gameHandler)

    // Setup Gin router
    router := gin.Default()

    // Register routes
    httpHandler.RegisterRoutes(router)
    wsHandler.RegisterRoutes(router)

    // Enhanced CORS middleware
    router.Use(func(c *gin.Context) {
        // Get allowed origins from environment or use default
        allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
        if allowedOrigin == "" {
            allowedOrigin = "*" // Default to allow all in development
        }
        
        c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    // Get port from environment variable or use default
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Default port
    }

    // Create server
    srv := &http.Server{
        Addr:    fmt.Sprintf(":%s", port),
        Handler: router,
    }

    // Start server in goroutine
    go func() {
        log.Printf("Server starting on port %s", port)
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