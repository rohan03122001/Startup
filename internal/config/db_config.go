package config

import (
	"fmt"
	"os"
	"strconv"
)

// GetDBConfig returns a database configuration using environment variables if available
func GetDBConfig() (*Database, error) {
	dbConfig := &Database{
		Host:     getEnv("DB_HOST", "localhost"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		User:     getEnv("DB_USER", "postgres"),
		DBName:   getEnv("DB_NAME", "quiz_app"),
	}

	// Parse port from environment variable
	portStr := getEnv("DB_PORT", "5434")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}
	dbConfig.Port = port

	return dbConfig, nil
}

// getEnv retrieves an environment variable value or returns a default value if not set
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}