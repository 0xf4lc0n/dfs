package di

import (
	"auth/database"
	"auth/services"
	"log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Logger *zap.Logger
var Db *gorm.DB
var MailService *services.MailService

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

	MailService = services.NewMailService(Logger)
}
