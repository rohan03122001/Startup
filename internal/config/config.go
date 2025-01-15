package config

type Config struct {
	Server   *Server
	Database *Database
}

type Server struct {
	Address string
}

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server:   &Server{},
		Database: &Database{},
	}

	cfg.Server.Address = ":8080"
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5434
	cfg.Database.User = "postgres"
	cfg.Database.Password = "postgres"
	cfg.Database.DBName = "quiz_app"

	return cfg, nil
}
