package config

import "github.com/spf13/viper"

type Config struct {
	Server   ServerConfig
	Database Database
	Redis    RedisConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLName  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadConfig()(*Config, error) {
	viper.SetConfigName("config") 
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")


	//setup defaults
	viper.SetDefault("server.port","8080")
	viper.SetDefault("server.mode","development")


	//Read the config
	if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

	//Unmarshal the file into struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

	return &config, nil

}