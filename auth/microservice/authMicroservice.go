package microservice

import (
	"log"

	"dfs/auth/config"
	"dfs/auth/controllers"
	"dfs/auth/database"
	"dfs/auth/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthMicroservice struct {
	config         *config.Config
	logger         *zap.Logger
	app            *fiber.App
	database       *gorm.DB
	usrRepo        *database.UserRepository
	vrfRepo        *database.VerificationRepository
	rpcClient      *services.RpcClient
	rpcServer      *services.RpcServer
	mail           *services.MailService
	authController *controllers.AuthController
}

func NewAuthMicroservice() *AuthMicroservice {
	cfg := config.Create()
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.EncoderConfig.FunctionKey = "func"
	logger, err := loggerCfg.Build()

	if err != nil {
		log.Fatalf("Cannot initialize zap logger. Reason: %s", err)
	}

	databaseService, err := database.Connect(cfg.DbConnectionString)

	if err != nil {
		log.Fatalf("Cannot initialize database service. Reason: %s", err)
	}

	app := fiber.New()
	rpcClient := services.NewRpcClient(logger)
	rpcServer := services.NewRpcServer(logger, databaseService)
	usrRepo := database.NewUserRepository(databaseService, logger)
	vrfRepo := database.NewVerificationRepository(databaseService, logger)
	mail := services.NewMailService(logger)
	authController := controllers.NewAuthController(logger, usrRepo, vrfRepo, mail, rpcClient)

	return &AuthMicroservice{config: cfg, logger: logger, app: app, database: databaseService, usrRepo: usrRepo,
		vrfRepo: vrfRepo, rpcClient: rpcClient, rpcServer: rpcServer, authController: authController}
}

func (ams *AuthMicroservice) Setup() {
	ams.app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	ams.authController.RegisterRoutes(ams.app)
}

func (ams *AuthMicroservice) Run() {
	go ams.rpcServer.RegisterGetUserDataByJwt()
	go ams.rpcServer.RegisterGetUserDataById()
	ams.app.Listen(":8080")
}

func (ams *AuthMicroservice) Cleanup() {
	ams.logger.Sync()
	ams.rpcServer.Close()
	ams.rpcClient.Close()
}
