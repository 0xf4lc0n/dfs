package microservice

import (
	"dfs/storageGateway/config"
	"dfs/storageGateway/controllers"
	"dfs/storageGateway/dtos"
	"dfs/storageGateway/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

type GatewayMicroservice struct {
	config            *config.Config
	logger            *zap.Logger
	nodes             *services.NodeService
	rpcClient         *services.RpcClient
	rpcServer         *services.RpcServer
	grpcClient        *services.GrpcStorageClient
	app               *fiber.App
	sessionStore      *session.Store
	gatewayController *controllers.GatewayController
}

func NewGatewayMicroservice() *GatewayMicroservice {
	cfg := config.Create()

	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.FunctionKey = "func"
	logger, err := loggerConfig.Build()

	if err != nil {
		log.Fatalf("Cannot initialize zap logger. Reason: %s", err)
	}

	app := fiber.New()
	store := session.New()

	nodeSrv := services.NewNodeService(logger)
	grpcClient := services.NewGrpcStorageClient(logger)

	rpcClient := services.NewRpcClient(logger)
	rpcServer := services.NewRpcServer(logger, nodeSrv, grpcClient)

	gatewayController := controllers.NewGatewayController(logger, store, nodeSrv)

	store.RegisterType(dtos.UserDto{})

	return &GatewayMicroservice{config: cfg, logger: logger, nodes: nodeSrv, grpcClient: grpcClient,
		rpcClient: rpcClient, rpcServer: rpcServer, app: app, sessionStore: store, gatewayController: gatewayController}
}

func (gm *GatewayMicroservice) Setup() {
	gm.app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	gm.app.Use(func(c *fiber.Ctx) error {
		//cookie := c.Cookies("jwt")
		//
		//userData := gm.rpcClient.GetUserDataByJwt(cookie)
		//
		//if userData == nil {
		//	return c.SendStatus(fiber.StatusUnauthorized)
		//}
		//
		//sess, err := gm.sessionStore.Get(c)
		//
		//if err != nil {
		//	gm.logger.Panic("Cannot get session", zap.Error(err))
		//}
		//
		//sess.Set("userData", userData)
		//
		//if err := sess.Save(); err != nil {
		//	gm.logger.Panic("Cannot save session", zap.Error(err))
		//}

		return c.Next()
	})

	gm.gatewayController.RegisterRoutes(gm.app)
}

func (gm *GatewayMicroservice) Run() {
	go gm.rpcServer.RegisterNodeMessages()
	go gm.rpcServer.RegisterGetFileByUniqueName()
	go gm.rpcServer.RegisterGetFileContentFromDisk()
	go gm.rpcServer.RegisterDeleteFileFromDisk()
	go gm.rpcServer.RegisterCreateHomeDirectory()
	go gm.rpcServer.RegisterGetFileById()
	go gm.rpcServer.RegisterGetOwnedFile()
	go gm.rpcServer.RegisterSaveFileOnDisk()

	gm.HandleInterrupt()

	if err := gm.app.Listen(gm.config.FullAddress); err != nil {
		gm.Cleanup()
		gm.logger.Panic("Cannot setup fiber listener", zap.Error(err))
	}
}

func (gm *GatewayMicroservice) HandleInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		gm.logger.Debug("Gracefully shutting down...")
		_ = gm.app.Shutdown()
	}()
}

func (gm *GatewayMicroservice) Cleanup() {
	gm.rpcServer.Close()
	gm.rpcClient.Close()

	if err := gm.logger.Sync(); err != nil {
		log.Printf("Cannot sync logger. Error: %s", err)
	}
}
