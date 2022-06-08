package microservice

import (
	"dfs/storageGateway/balancer"
	"dfs/storageGateway/config"
	"dfs/storageGateway/controllers"
	"dfs/storageGateway/dtos"
	"dfs/storageGateway/node"
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
	nodes             *node.ActiveNodes
	balancer          *balancer.RoundRobin
	rpcClient         *services.RpcClient
	rpcServer         *services.RpcServer
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

	activeNodes := node.NewNodes(logger)
	rrBalancer := balancer.New(activeNodes)

	rpcClient := services.NewRpcClient(logger)
	rpcServer := services.NewRpcServer(logger, activeNodes)

	gatewayController := controllers.NewGatewayController(logger, store, rrBalancer)

	store.RegisterType(dtos.UserDto{})

	return &GatewayMicroservice{config: cfg, logger: logger, nodes: activeNodes, balancer: rrBalancer,
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
