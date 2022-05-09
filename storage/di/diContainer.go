package di

import (
	"log"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitializeServices() {
	loggerService, err := zap.NewDevelopment()

	if err != nil {
		log.Fatalf("Cannot initialize zap logger. Reason: %s", err)
	}

	Logger = loggerService
}
