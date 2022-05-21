package microservice

import (
	"log"

	"dfs/auth/controllers"
	"dfs/auth/database"
	"dfs/auth/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthMicroservice struct {
	logger         *zap.Logger
	app            *fiber.App
	database       *gorm.DB
	rpcClient      *services.RpcClient
	rpcServer      *services.RpcServer
	mail           *services.MailService
	authController *controllers.AuthController
}

func NewAuthMicroservice() *AuthMicroservice {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.FunctionKey = "func"
	logger, err := config.Build()

	if err != nil {
		log.Fatalf("Cannot initialize zap logger. Reason: %s", err)
	}

	databaseService, err := database.Connect()

	if err != nil {
		log.Fatalf("Cannot initialize database service. Reason: %s", err)
	}

	app := fiber.New()
	rpcClient := services.NewRpcClient(logger)
	rpcServer := services.NewRpcServer(logger, databaseService)
	mail := services.NewMailService(logger)
	authController := controllers.NewAuthController(logger, databaseService, mail, rpcClient)

	return &AuthMicroservice{logger: logger, app: app, database: databaseService, rpcClient: rpcClient, rpcServer: rpcServer, authController: authController}
}

func (ams *AuthMicroservice) Setup() {
	ams.app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	ams.authController.RegisterRoutes(ams.app)
}

func (ams *AuthMicroservice) Run() {
	go ams.rpcServer.RegisterGetUserHomeDirectory()
	go ams.rpcServer.RegisterValidateJwt()
	go ams.rpcServer.RegisterGetUserDataByJwt()
	go ams.rpcServer.RegisterGetUserDataById()
	ams.app.Listen(":8080")
}

func (ams *AuthMicroservice) Cleanup() {
	ams.logger.Sync()
	ams.rpcServer.Close()
	ams.rpcClient.Close()
}
