package microservice

import (
	"log"

	"dfs/storage/controllers"
	"dfs/storage/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
)

type StorageMicroservice struct {
	logger         *zap.Logger
	app            *fiber.App
	rpcClient      *services.RpcClient
	rpcServer      *services.RpcServer
	fileController *controllers.FileController
}

func NewStorageMicroservice() *StorageMicroservice {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.FunctionKey = "func"
	logger, err := config.Build()

	if err != nil {
		log.Fatalf("Cannot initialize zap logger. Reason: %s", err)
	}

	app := fiber.New()
	rpcClient := services.NewRpcClient(logger)
	rpcServer := services.NewRpcServer(logger)
	fileController := controllers.NewFileController(logger, rpcClient)

	return &StorageMicroservice{logger: logger, app: app, rpcClient: rpcClient, rpcServer: rpcServer, fileController: fileController}
}

func (sms *StorageMicroservice) Setup() {
	sms.app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	sms.app.Use(func(c *fiber.Ctx) error {
		cookie := c.Cookies("jwt")

		if sms.rpcClient.IsAuthenticated(cookie) == false {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.Next()
	})

	sms.fileController.RegisterRoutes(sms.app)
}

func (sms *StorageMicroservice) Run() {
	go sms.rpcServer.RegisterCreateHomeDirectory()
	sms.app.Listen(":8081")
}

func (sms *StorageMicroservice) Cleanup() {
	sms.logger.Sync()
	sms.rpcServer.Close()
	sms.rpcClient.Close()
}
