package config

import "os"

type Config struct {
	Port         string
	DBPath       string
	AdminPass    string
	SessionSecret string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "blog.db"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "mago-default-secret-change-in-production"
	}

	return &Config{
		Port:         port,
		DBPath:       dbPath,
		AdminPass:    os.Getenv("ADMIN_PASSWORD"),
		SessionSecret: sessionSecret,
	}
}
