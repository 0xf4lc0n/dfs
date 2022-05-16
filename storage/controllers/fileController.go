package controllers

import (
	"dfs/storage/services"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FileController struct {
	logger    *zap.Logger
	rpcClient *services.RpcClient
}

const STORAGE_PATH = `C:\Users\Falcon\Desktop\Files`

func NewFileController(logger *zap.Logger, rpcClient *services.RpcClient) *FileController {
	return &FileController{logger: logger, rpcClient: rpcClient}
}

func (fc *FileController) RegisterRoutes(app *fiber.App) {
	app.Post("/api/file", fc.uploadFile)
	app.Get("/api/file/:fileName", fc.downloadFile)
}

func (fc *FileController) uploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	jwt := c.Cookies("jwt")
	homeDir := fc.rpcClient.GetHomeDirectory(jwt)

	if homeDir == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	fileName := uuid.New().String()
	fullFilePath := path.Join(STORAGE_PATH, homeDir, fileName)
	return c.SaveFile(file, fullFilePath)
}

func (fc *FileController) downloadFile(c *fiber.Ctx) error {
	fileName := c.Params("fileName")
	fullFilePath := path.Join(STORAGE_PATH, fileName)
	return c.Download(fullFilePath)
}
