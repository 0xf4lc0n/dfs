package database

import (
	"dfs/auth/models"
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(connectionString string) (*gorm.DB, error) {
	connection, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		return nil, errors.New("cannot connect to the database")
	}

	connection.AutoMigrate(&models.User{})
	connection.AutoMigrate(&models.VerificationData{})

	return connection, nil
}
