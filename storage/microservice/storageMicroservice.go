package microservice

import (
	"log"

	"dfs/storage/controllers"
	"dfs/storage/database"
	"dfs/storage/dtos"
	"dfs/storage/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StorageMicroservice struct {
	logger         *zap.Logger
	app            *fiber.App
	store          *session.Store
	database       *gorm.DB
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

	databaseService, err := database.Connect()

	if err != nil {
		log.Fatalf("Cannot initialize database service. Reason: %s", err)
	}

	app := fiber.New()
	rpcClient := services.NewRpcClient(logger)
	rpcServer := services.NewRpcServer(logger)
	store := session.New()
	fileController := controllers.NewFileController(logger, rpcClient, databaseService, store)

	store.RegisterType(dtos.User{})

	return &StorageMicroservice{logger: logger, app: app, store: store, database: databaseService, rpcClient: rpcClient, rpcServer: rpcServer, fileController: fileController}
}

func (sms *StorageMicroservice) Setup() {
	sms.app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	sms.app.Use(func(c *fiber.Ctx) error {
		cookie := c.Cookies("jwt")

		userData := sms.rpcClient.GetUserData(cookie)

		if userData == nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		sess, err := sms.store.Get(c)

		if err != nil {
			sms.logger.Panic("Cannot get session", zap.Error(err))
		}

		sess.Set("userData", userData)

		if err := sess.Save(); err != nil {
			sms.logger.Panic("Cannot save session", zap.Error(err))
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
