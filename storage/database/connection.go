package database

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	err := godotenv.Load(".env")

	if err != nil {
		return nil, errors.New("Cannot load .env file")
	}

	connectionString := os.Getenv("DB_CONNECTION_STRING")
	connection, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		return nil, errors.New("Could not connect to the database")
	}

	return connection, nil
}
