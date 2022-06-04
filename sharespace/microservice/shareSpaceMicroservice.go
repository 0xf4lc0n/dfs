package microservice

import (
	"dfs/sharespace/config"
	"dfs/sharespace/controllers"
	"dfs/sharespace/database"
	"dfs/sharespace/dtos"
	"dfs/sharespace/services"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ShareSpaceMicroservice struct {
	config               *config.Config
	logger               *zap.Logger
	app                  *fiber.App
	store                *session.Store
	database             *gorm.DB
	ssRepository         *database.ShareSpaceRepository
	rpcClient            *services.RpcClient
	rpcServer            *services.RpcServer
	shareSpaceController *controllers.ShareSpaceController
}

func NewShareSpaceMicroservice() *ShareSpaceMicroservice {
	cfg := config.Create()
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.FunctionKey = "func"
	logger, err := loggerConfig.Build()

	if err != nil {
		log.Fatalf("Cannot initialize zap logger. Reason: %s", err)
	}

	databaseService, err := database.Connect(cfg.DbConnectionString)

	if err != nil {
		log.Fatalf("Cannot initialize database service. Reason: %s", err)
	}

	app := fiber.New()
	rpcClient := services.NewRpcClient(logger)
	rpcServer := services.NewRpcServer(logger)
	ssRepository := database.NewShareSpaceRepository(logger, databaseService, rpcClient)
	store := session.New()
	shareSpaceController := controllers.NewShareSpaceController(logger, rpcClient, ssRepository, store)
	store.RegisterType(dtos.UserDto{})

	return &ShareSpaceMicroservice{config: cfg, logger: logger, app: app, store: store, database: databaseService,
		ssRepository: ssRepository, rpcClient: rpcClient, rpcServer: rpcServer, shareSpaceController: shareSpaceController}
}

func (sms *ShareSpaceMicroservice) Setup() {
	sms.app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	sms.app.Use(func(c *fiber.Ctx) error {
		cookie := c.Cookies("jwt")

		userData := sms.rpcClient.GetUserDataByJwt(cookie)

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

	sms.shareSpaceController.RegisterRoutes(sms.app)
}

func (sms *ShareSpaceMicroservice) Run() {
	// go sms.rpcServer.RegisterCreateHomeDirectory()
	sms.app.Listen(":8083")
}

func (sms *ShareSpaceMicroservice) Cleanup() {
	sms.logger.Sync()
	sms.rpcServer.Close()
	sms.rpcClient.Close()
}
