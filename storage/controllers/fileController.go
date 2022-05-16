package controllers

import (
	"dfs/auth/models"
	"dfs/storage/services"
	"path"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FileController struct {
	logger    *zap.Logger
	rpcClient *services.RpcClient
	database  *gorm.DB
}

const STORAGE_PATH = `C:\Users\Falcon\Desktop\Files`

func NewFileController(logger *zap.Logger, rpcClient *services.RpcClient, database *gorm.DB) *FileController {
	return &FileController{logger: logger, rpcClient: rpcClient, database: database}
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
	userData := fc.rpcClient.GetUserData(jwt)

	if userData == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	fileName := uuid.New().String()
	fullFilePath := path.Join(STORAGE_PATH, userData.HomeDirectory, fileName)

	saveFileResult := c.SaveFile(file, fullFilePath)

	if saveFileResult == nil {
		fileEntry := &models.File{
			UniqueName:   fileName,
			Name:         file.Filename,
			CreationDate: time.Now(),
			OwnerId:      userData.Id,
		}

		fc.database.Create(&fileEntry)
	}

	return saveFileResult
}

func (fc *FileController) downloadFile(c *fiber.Ctx) error {
	fileName := c.Params("fileName")
	fullFilePath := path.Join(STORAGE_PATH, fileName)
	return c.Download(fullFilePath)
}
