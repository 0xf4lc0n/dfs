package services

import (
	"auth/database"
	"log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Logger *zap.Logger
var Db *gorm.DB

func InitializeServices() {
	loggerService, err := zap.NewDevelopment()

	if err != nil {
		log.Fatalf("Cannot initialize zap logger. Reason: %s", err)
	}

	databaseService, err := database.Connect()

	if err != nil {
		log.Fatalf("Cannot initialize database service. Reason: %s", err)
	}

	Logger = loggerService
	Db = databaseService
}
