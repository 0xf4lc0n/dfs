package controllers

import (
	"dfs/share/database"
	"dfs/share/dtos"
	"dfs/share/services"
	"encoding/base64"
	"fmt"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.uber.org/zap"
)

type ShareController struct {
	log    *zap.Logger
	rpc    *services.RpcClient
	store  *session.Store
	shRepo *database.ShareRepository
}

func NewShareController(logger *zap.Logger, rpcClient *services.RpcClient, store *session.Store, shRepo *database.ShareRepository) *ShareController {
	return &ShareController{log: logger, rpc: rpcClient, store: store, shRepo: shRepo}
}

func (sc *ShareController) RegisterRoutes(app *fiber.App) {
	app.Post("/api/share", sc.shareFile)
	app.Delete("/api/share", sc.unshareFile)
	app.Get("/api/share", sc.getFilesSharedForUser)
	app.Get("/api/share/me", sc.getFilesSharedByUser)
	app.Get("/api/share/:uniqueFileName", sc.downloadSharedFile)
}

func (sc *ShareController) shareFile(c *fiber.Ctx) error {
	shareDto := new(dtos.ShareDto)

	if err := c.BodyParser(&shareDto); err != nil {
		sc.log.Warn("Cannot parse share data", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if shareDto.SharedById == shareDto.SharedToId {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot share file for yourself"})
	}

	sess, err := sc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		sc.log.Panic("Cannot get session", zap.Error(err))
	}

	sharedBy := sess.Get("userData").(dtos.UserDto)

	if shareDto.SharedById != sharedBy.Id {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot share as other user"})
	}

	sharedFor := sc.rpc.GetUserDataById(shareDto.SharedToId)

	if sharedFor == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot share file"})
	}

	sharedFile := sc.rpc.GetOwnedFile(&dtos.OwnedFileDto{OwnerId: shareDto.SharedById, FileId: shareDto.FileId})

	if sharedFile == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot share nonexisting file"})
	}

	if sc.shRepo.CreateShareFileEntry(sharedFile.Id, sharedFor.Id, sharedBy.Id, shareDto.ExpirationTime) == false {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot share file"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (sc *ShareController) getFilesSharedForUser(c *fiber.Ctx) error {
	sess, err := sc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		sc.log.Panic("Cannot get session", zap.Error(err))
	}

	user := sess.Get("userData").(dtos.UserDto)

	shares := sc.shRepo.GetSharedForUserFilesEntries(user.Id)

	var files []dtos.SharedFileDto

	for _, share := range shares {
		file := sc.rpc.GetFileById(share.FileId)
		fileOwner := sc.rpc.GetUserDataById(file.OwnerId)
		sharedBy := sc.rpc.GetUserDataById(share.SharedById)

		sharedFile := dtos.SharedFileDto{
			Name:        file.Name,
			UniqueName:  file.UniqueName,
			Owner:       fileOwner.Name,
			SharedBy:    sharedBy.Name,
			AvailableTo: share.ExpirationTime,
		}

		files = append(files, sharedFile)
	}

	return c.JSON(files)
}

func (sc *ShareController) getFilesSharedByUser(c *fiber.Ctx) error {
	sess, err := sc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		sc.log.Panic("Cannot get session", zap.Error(err))
	}

	user := sess.Get("userData").(dtos.UserDto)

	shares := sc.shRepo.GetSharedByUserFilesEntries(user.Id)

	var files []dtos.SharedForDto

	for _, share := range shares {
		file := sc.rpc.GetOwnedFile(&dtos.OwnedFileDto{FileId: share.FileId, OwnerId: user.Id})
		sharedForEntries := sc.shRepo.GetSharedEntriesByFileId(file.Id)

		var sharedForUsers []string

		for _, sharedFor := range sharedForEntries {
			user := sc.rpc.GetUserDataById(sharedFor.SharedForId)
			userName := fmt.Sprintf("%s (%s)", user.Name, user.Email)

			sharedForUsers = append(sharedForUsers, userName)
		}

		sharedFile := dtos.SharedForDto{
			Name:        file.Name,
			UniqueName:  file.UniqueName,
			SharedFor:   sharedForUsers,
			AvailableTo: share.ExpirationTime,
		}

		files = append(files, sharedFile)
	}

	return c.JSON(files)
}

func (sc *ShareController) unshareFile(c *fiber.Ctx) error {
	unshareDto := new(dtos.UnshareDto)

	if err := c.BodyParser(&unshareDto); err != nil {
		sc.log.Warn("Cannot parse unshare data", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	sess, err := sc.store.Get(c)
	defer sess.Destroy()

	if err != nil {
		sc.log.Panic("Cannot get session", zap.Error(err))
	}

	user := sess.Get("userData").(dtos.UserDto)

	if sc.shRepo.DeleteShareFileEntry(unshareDto.FileId, unshareDto.SharedForId, user.Id) == false {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot unshare file"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (sc *ShareController) downloadSharedFile(ctx *fiber.Ctx) error {
	fileUniqueName := ctx.Params("uniqueFileName")

	sess, err := sc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		sc.log.Panic("Cannot get session", zap.Error(err))
	}

	userData := sess.Get("userData").(dtos.UserDto)

	file := sc.rpc.GetFileByUniqueName(fileUniqueName)

	if file == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot download non-shared file"})
	}

	sharedFileEntry := sc.shRepo.GetSharedForFileEntry(file.Id, userData.Id)

	if sharedFileEntry == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot download non-shared file"})
	}

	fileOwner := sc.rpc.GetUserDataById(file.OwnerId)

	if fileOwner == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot download non-shared file"})
	}

	readPath := path.Join(fileOwner.HomeDirectory, file.UniqueName)
	decryptionKey, err := base64.StdEncoding.DecodeString(fileOwner.CryptKey)

	if err != nil {
		sc.log.Error("Cannot decode decryption key", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot download shared file"})
	}

	readFileDto := dtos.ReadFileDto{ReadPath: readPath, DecryptionKey: decryptionKey}

	fileContent := sc.rpc.ReadFileFromDisk(readFileDto)

	if fileContent == nil {
		sc.log.Error("Cannot read shared file from disk", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot download shared file"})
	}

	return ctx.Send(fileContent)
}
