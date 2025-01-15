// internal/config/config.go

package config

import (
	"github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
}

type ServerConfig struct {
    Port string
    Mode string
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")

    // Default values
    viper.SetDefault("server.port", "8080")
    viper.SetDefault("server.mode", "development")
    
    viper.SetDefault("database.host", "localhost")
    viper.SetDefault("database.port", "5432")
    viper.SetDefault("database.sslmode", "disable")

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}