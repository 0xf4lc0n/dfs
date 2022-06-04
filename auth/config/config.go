package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	DbConnectionString string
	JwtSecretKey       string
}

func Create() *Config {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Cannot load .env file")
	}

	cfg := &Config{
		DbConnectionString: os.Getenv("DB_CONNECTION_STRING"),
		JwtSecretKey:       os.Getenv("JWT_SECRET_KEY"),
	}

	return cfg
}
