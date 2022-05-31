package database

import (
	"dfs/sharespace/models"
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(connectionString string) (*gorm.DB, error) {
	connection, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		return nil, errors.New("Could not connect to the database")
	}

	connection.AutoMigrate(&models.ShareSpace{})
	connection.AutoMigrate(&models.ShareSpaceMember{})
	connection.AutoMigrate(&models.ShareSpaceFile{})

	return connection, nil
}
