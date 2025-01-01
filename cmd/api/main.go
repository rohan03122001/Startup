package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/rohan03122001/quizzing/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil{
		log.Fatalf("Failed to Load %v", err)
	}

	r:= gin.Default()

	r.GET("/health", )

	log.Printf("Server Starting at port %s", cfg.Server.Port)
	if err:= r.Run(":"+cfg.Server.Port); err != nil{
		log.Fatal("Error Running the server")
	}
}
