package database

import (
	"dfs/storage/models"
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(connectionString string) (*gorm.DB, error) {
	connection, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		return nil, errors.New("could not connect to the database")
	}

	connection.AutoMigrate(&models.File{})

	return connection, nil
}
