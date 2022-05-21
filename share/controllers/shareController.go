package controllers

import (
	"dfs/share/dtos"
	"dfs/share/models"
	"dfs/share/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ShareController struct {
	logger    *zap.Logger
	rpcClient *services.RpcClient
	database  *gorm.DB
	store     *session.Store
}

func NewShareController(logger *zap.Logger, rpcClient *services.RpcClient, database *gorm.DB, store *session.Store) *ShareController {
	return &ShareController{logger: logger, rpcClient: rpcClient, database: database, store: store}
}

func (sc *ShareController) RegisterRoutes(app *fiber.App) {
	app.Post("/api/share", sc.shareFile)
	app.Delete("/api/share", sc.unshareFile)
	app.Get("/api/share", sc.getSharedFiles)
}

func (sc *ShareController) shareFile(c *fiber.Ctx) error {
	shareDto := new(dtos.ShareDto)

	if err := c.BodyParser(&shareDto); err != nil {
		sc.logger.Warn("Cannot parse register data", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if shareDto.SharedById == shareDto.UserId {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot share file for yourself"})
	}

	sess, err := sc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		sc.logger.Panic("Cannot get session", zap.Error(err))
	}

	sharedBy := sess.Get("userData").(dtos.UserDto)

	sharedFor := sc.rpcClient.GetUserDataById(shareDto.UserId)

	if sharedFor == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	sharedFile := sc.rpcClient.GetOwnedFile(&dtos.OwnedFileDto{OwnerId: shareDto.SharedById, FileId: shareDto.FileId})

	if sharedFile == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	share := &models.Share{FileId: sharedFile.Id, UserId: sharedFor.Id, SharedById: sharedBy.Id, ExpirationTime: shareDto.ExpirationTime}

	sc.database.Create(&share)

	return c.SendStatus(fiber.StatusOK)
}

func (sc *ShareController) getSharedFiles(c *fiber.Ctx) error {
	sess, err := sc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		sc.logger.Panic("Cannot get session", zap.Error(err))
	}

	user := sess.Get("userData").(dtos.UserDto)

	shares := new([]models.Share)
	sc.database.Where("user_id = ?", user.Id).Find(&shares)

	files := []dtos.FileDto{}

	for _, share := range *shares {
		file := sc.rpcClient.GetFileById(share.FileId)

		files = append(files, *file)
	}

	return c.JSON(files)
}

func (sc *ShareController) unshareFile(c *fiber.Ctx) error {
	unshareDto := new(dtos.UnshareDto)

	if err := c.BodyParser(&unshareDto); err != nil {
		sc.logger.Warn("Cannot parse register data", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	sess, err := sc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		sc.logger.Panic("Cannot get session", zap.Error(err))
	}

	user := sess.Get("userData").(dtos.UserDto)

	fileDto := sc.rpcClient.GetOwnedFile(&dtos.OwnedFileDto{FileId: unshareDto.FileId, OwnerId: user.Id})

	if fileDto == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	share := &models.Share{FileId: fileDto.Id, UserId: unshareDto.UserId}
	sc.database.Delete(&share)

	return c.SendStatus(fiber.StatusOK)
}
