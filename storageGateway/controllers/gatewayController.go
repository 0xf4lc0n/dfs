package controllers

import (
	"dfs/storageGateway/services"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.uber.org/zap"
)

type GatewayController struct {
	logger *zap.Logger
	store  *session.Store
	nodes  *services.NodeService
}

func NewGatewayController(log *zap.Logger, store *session.Store, nodes *services.NodeService) *GatewayController {
	return &GatewayController{logger: log, store: store, nodes: nodes}
}

func (gc *GatewayController) RegisterRoutes(app *fiber.App) {
	app.Post("/api/file", gc.uploadFile)
	app.Get("/api/file/:fileUniqueName", gc.downloadFile)
	app.Get("/api/file", gc.getFiles)
	app.Delete("/api/file/:fileUniqueName", gc.deleteFile)
}

func (gc *GatewayController) uploadFile(ctx *fiber.Ctx) error {
	activeNodes := gc.nodes.GetNodes()

	for _, n := range *activeNodes {
		gc.logger.Debug("Gateway picked node", zap.String("NodeAddress", n.IpAddress))

		url := fmt.Sprintf("http://%s:%d/api/file", n.IpAddress, n.Port)

		if err := proxy.Do(ctx, url); err != nil {
			gc.logger.Error("Error during proxying request to selected node", zap.String("NodeAddress", n.IpAddress),
				zap.Error(err))
		}
	}

	ctx.Response().Header.Del(fiber.HeaderServer)

	return ctx.SendStatus(fiber.StatusOK)
}

func (gc *GatewayController) deleteFile(ctx *fiber.Ctx) error {
	fileUniqueName := ctx.Params("fileUniqueName")
	activeNodes := gc.nodes.GetNodes()

	for _, n := range *activeNodes {
		gc.logger.Debug("Gateway picked node", zap.String("NodeAddress", n.IpAddress))

		url := fmt.Sprintf("http://%s:%d/api/file/%s", n.IpAddress, n.Port, fileUniqueName)

		if err := proxy.Do(ctx, url); err != nil {
			gc.logger.Error("Error during proxying request to selected node", zap.String("NodeAddress", n.IpAddress),
				zap.Error(err))
		}
	}

	ctx.Response().Header.Del(fiber.HeaderServer)

	return ctx.SendStatus(fiber.StatusOK)
}

func (gc *GatewayController) downloadFile(ctx *fiber.Ctx) error {
	fileUniqueName := ctx.Params("fileUniqueName")
	n := gc.nodes.Next()

	gc.logger.Debug("Gateway picked node", zap.String("NodeAddress", n.IpAddress))

	url := fmt.Sprintf("http://%s:%d/api/file/%s", n.IpAddress, n.Port, fileUniqueName)

	if err := proxy.Do(ctx, url); err != nil {
		gc.logger.Error("Error during proxying request to selected node", zap.String("NodeAddress", n.IpAddress),
			zap.Error(err))
	}

	ctx.Response().Header.Del(fiber.HeaderServer)

	return ctx.SendStatus(fiber.StatusOK)
}

func (gc *GatewayController) getFiles(ctx *fiber.Ctx) error {
	n := gc.nodes.Next()

	gc.logger.Debug("Gateway picked node", zap.String("NodeAddress", n.IpAddress))

	url := fmt.Sprintf("http://%s:%d/api/file", n.IpAddress, n.Port)

	if err := proxy.Do(ctx, url); err != nil {
		gc.logger.Error("Error during proxying request to selected node", zap.String("NodeAddress", n.IpAddress),
			zap.Error(err))
	}

	ctx.Response().Header.Del(fiber.HeaderServer)

	return ctx.SendStatus(fiber.StatusOK)
}
