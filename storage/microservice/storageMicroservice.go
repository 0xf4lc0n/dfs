package microservice

import (
	"dfs/proto"
	"dfs/storage/config"
	"dfs/storage/controllers"
	"dfs/storage/database"
	"dfs/storage/dtos"
	"dfs/storage/node"
	"dfs/storage/services"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"log"
	"net"
	"os"
	"os/signal"
)

type StorageMicroservice struct {
	uuid           uuid.UUID
	config         *config.Config
	logger         *zap.Logger
	app            *fiber.App
	store          *session.Store
	database       *gorm.DB
	rpcClient      *services.RpcClient
	grpcServer     *services.GRpcStorageServer
	fileService    *services.FileService
	storageRpo     *database.StorageRepository
	fileController *controllers.FileController
}

func NewStorageMicroservice() *StorageMicroservice {
	uid := uuid.New()
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
	fileService := services.NewFileService(cfg, logger)
	rpcClient := services.NewRpcClient(logger, uid)
	store := session.New()
	storageRepository := database.NewStorageRepository(logger, databaseService)
	grpcServer := services.NewGrpcStorageServer(logger, fileService, storageRepository)
	fileController := controllers.NewFileController(cfg, logger, rpcClient, store, storageRepository, fileService)

	store.RegisterType(dtos.User{})

	return &StorageMicroservice{uuid: uid, config: cfg, logger: logger, app: app, store: store,
		database: databaseService, rpcClient: rpcClient, fileService: fileService, storageRpo: storageRepository,
		grpcServer: grpcServer, fileController: fileController}
}

func (sms *StorageMicroservice) Setup() {
	sms.app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	fileApi := sms.app.Group("/api/file", func(c *fiber.Ctx) error {
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

	sms.fileController.RegisterRoutes(&fileApi)
}

func (sms *StorageMicroservice) Run() {
	sms.rpcClient.SendNodeMessage(node.CreateRegisterNodeMessage(&node.Node{
		Uuid:      sms.uuid,
		IpAddress: sms.config.IpAddress,
		Port:      sms.config.Port,
		GrpcPort:  sms.config.GRpcPort,
	}))

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", sms.config.IpAddress, sms.config.GRpcPort))

	if err != nil {
		sms.Cleanup()
		sms.logger.Panic("Cannot create listener for GRPC server", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	proto.RegisterStorageServer(grpcServer, sms.grpcServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			sms.logger.Panic("Cannot serve GRPC server", zap.Error(err))
		}
	}()

	//go sms.rpcServer.RegisterCreateHomeDirectory()
	//go sms.rpcServer.RegisterGetOwnedFile()
	//go sms.rpcServer.RegisterGetFileById()
	//go sms.rpcServer.RegisterGetFileByUniqueName()
	//go sms.rpcServer.RegisterSaveFileOnDisk()
	//go sms.rpcServer.RegisterDeleteFileFromDisk()
	//go sms.rpcServer.RegisterGetFileContentFromDisk()

	sms.HandleInterrupt()

	if err := sms.app.Listen(sms.config.FullAddress); err != nil {
		sms.Cleanup()
		sms.logger.Panic("Cannot setup fiber listener", zap.Error(err))
	}
}

func (sms *StorageMicroservice) HandleInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		sms.logger.Debug("Gracefully shutting down...")
		_ = sms.app.Shutdown()
	}()
}

func (sms *StorageMicroservice) Cleanup() {
	sms.rpcClient.SendNodeMessage(node.CreateDeregisterNodeMessage(&node.Node{
		Uuid:      sms.uuid,
		IpAddress: sms.config.IpAddress,
		Port:      sms.config.Port,
		GrpcPort:  sms.config.GRpcPort,
	}))

	//sms.rpcServer.Close()
	sms.rpcClient.Close()

	if err := sms.logger.Sync(); err != nil {
		log.Printf("Cannot sync logger. Error: %s", err)
	}

}
