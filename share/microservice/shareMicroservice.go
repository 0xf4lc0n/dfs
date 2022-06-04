package microservice

import (
	"dfs/share/config"
	"log"

	"dfs/share/controllers"
	"dfs/share/database"
	"dfs/share/dtos"
	"dfs/share/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ShareMicroservice struct {
	config          *config.Config
	logger          *zap.Logger
	app             *fiber.App
	store           *session.Store
	database        *gorm.DB
	shareRepository *database.ShareRepository
	rpcClient       *services.RpcClient
	rpcServer       *services.RpcServer
	fileController  *controllers.ShareController
}

func NewShareMicroservice() *ShareMicroservice {
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
	rpcServer := services.NewRpcServer(logger)
	shareRepo := database.NewShareRepository(logger, databaseService)
	store := session.New()
	fileController := controllers.NewShareController(logger, rpcClient, store, shareRepo)
	store.RegisterType(dtos.UserDto{})

	return &ShareMicroservice{config: cfg, logger: logger, app: app, store: store, database: databaseService,
		rpcClient: rpcClient, rpcServer: rpcServer, shareRepository: shareRepo, fileController: fileController}
}

func (sms *ShareMicroservice) Setup() {
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

	sms.fileController.RegisterRoutes(sms.app)
}

func (sms *ShareMicroservice) Run() {
	sms.app.Listen(":8082")
}

func (sms *ShareMicroservice) Cleanup() {
	sms.logger.Sync()
	sms.rpcServer.Close()
	sms.rpcClient.Close()
}
