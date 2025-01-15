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
    Port string `mapstructure:"port"`
    Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     string `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    DBName   string `mapstructure:"dbname"`
    SSLMode  string `mapstructure:"sslmode"`
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")

    // Set defaults
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