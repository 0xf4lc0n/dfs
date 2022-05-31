package controllers

import (
	"dfs/storage/config"
	"dfs/storage/dtos"
	"dfs/storage/models"
	"dfs/storage/services"
	"path"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FileController struct {
	config    *config.Config
	logger    *zap.Logger
	rpcClient *services.RpcClient
	database  *gorm.DB
	store     *session.Store
}

func NewFileController(cfg *config.Config, logger *zap.Logger, rpcClient *services.RpcClient, database *gorm.DB, store *session.Store) *FileController {
	return &FileController{config: cfg, logger: logger, rpcClient: rpcClient, database: database, store: store}
}

func (fc *FileController) RegisterRoutes(app *fiber.App) {
	app.Post("/api/file", fc.uploadFile)
	app.Get("/api/file/:fileUniqueName", fc.downloadFile)
	app.Get("/api/file", fc.getUserFiles)
}

func (fc *FileController) uploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	sess, err := fc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		fc.logger.Panic("Cannot get session", zap.Error(err))
	}

	userData := sess.Get("userData").(dtos.User)

	fileName := uuid.New().String()
	fullFilePath := path.Join(fc.config.FileStoragePath, userData.HomeDirectory, fileName)

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
	fileUniqueName := c.Params("fileUniqueName")

	sess, err := fc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		fc.logger.Panic("Cannot get session", zap.Error(err))
	}

	userData := sess.Get("userData").(dtos.User)
	var file models.File

	if err = fc.database.Where("owner_id = ? AND unique_name = ?", userData.Id, fileUniqueName).First(&file).Error; err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	fileName := file.UniqueName
	fullFilePath := path.Join(fc.config.FileStoragePath, userData.HomeDirectory, fileName)

	fc.logger.Debug("File path", zap.String("FilePath", fullFilePath))

	return c.Download(fullFilePath)
}

func (fc *FileController) getUserFiles(c *fiber.Ctx) error {
	sess, err := fc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		fc.logger.Panic("Cannot get session", zap.Error(err))
	}

	userData := sess.Get("userData").(dtos.User)
	fc.logger.Debug("User id:", zap.Uint("userId", userData.Id))

	userId := userData.Id

	files := new([]models.File)

	fc.database.Where("owner_id = ?", userId).Find(&files)

	return c.JSON(files)
}
