package controllers

import (
	"dfs/storageGateway/balancer"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.uber.org/zap"
)

type GatewayController struct {
	logger   *zap.Logger
	store    *session.Store
	balancer *balancer.RoundRobin
}

func NewGatewayController(log *zap.Logger, store *session.Store, rrBalancer *balancer.RoundRobin) *GatewayController {
	return &GatewayController{logger: log, store: store, balancer: rrBalancer}
}

func (gc *GatewayController) RegisterRoutes(app *fiber.App) {
	app.Post("/api/file", gc.uploadFile)
	app.Get("/api/file/:fileUniqueName", gc.downloadFile)
	app.Get("/api/file", gc.getFiles)
	app.Delete("/api/file/:fileUniqueName", gc.deleteFile)
}

func (gc *GatewayController) uploadFile(ctx *fiber.Ctx) error {
	pickedNode := gc.balancer.Next()

	gc.logger.Debug("Gateway picked node", zap.String("NodeAddress", pickedNode))

	return ctx.SendStatus(fiber.StatusOK)
}

func (gc *GatewayController) deleteFile(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}

func (gc *GatewayController) downloadFile(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}

func (gc *GatewayController) getFiles(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}
