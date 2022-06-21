package controllers

import (
	"dfs/storage/config"
	"dfs/storage/database"
	"dfs/storage/dtos"
	"dfs/storage/services"
	"encoding/base64"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io/ioutil"
	"path"
)

type FileController struct {
	cfg        *config.Config
	log        *zap.Logger
	rpc        *services.RpcClient
	store      *session.Store
	storageRpo *database.StorageRepository
	fileSrv    *services.FileService
}

func NewFileController(cfg *config.Config, log *zap.Logger, rpc *services.RpcClient, store *session.Store,
	storageRpo *database.StorageRepository, fileSrv *services.FileService) *FileController {
	return &FileController{cfg: cfg, log: log, rpc: rpc, store: store, storageRpo: storageRpo, fileSrv: fileSrv}
}

func (fc *FileController) RegisterRoutes(app *fiber.Router) {
	(*app).Post("/", fc.uploadFile)
	(*app).Get("/:fileUniqueName", fc.downloadFile)
	(*app).Get("/", fc.getUserFiles)
	(*app).Delete("/:fileUniqueName", fc.deleteFile)
}

func (fc *FileController) uploadFile(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	sess, err := fc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		fc.log.Panic("Cannot get session", zap.Error(err))
	}

	userData := sess.Get("userData").(dtos.User)

	fileUniqueName := uuid.New().String()
	fileSavePath := path.Join(userData.HomeDirectory, fileUniqueName)

	file, err := fileHeader.Open()

	if err != nil {
		fc.log.Error("Cannot open file", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	defer file.Close()

	fileContent, err := ioutil.ReadAll(file)

	if err != nil {
		fc.log.Error("Cannot read file", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	encryptionKey, err := base64.StdEncoding.DecodeString(userData.CryptKey)

	if err != nil {
		fc.log.Error("Cannot decode encryption key", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	if fc.fileSrv.EncryptAndSaveFile(fileSavePath, fileContent, encryptionKey) == false {
		fc.log.Error("Cannot save file on the disk")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	if fc.storageRpo.CheckIfOwnedFileExistByName(fileHeader.Filename, userData.Id) {
		return ctx.SendStatus(fiber.StatusCreated)
	}

	if fc.storageRpo.CreateFile(fileUniqueName, fileHeader.Filename, userData.Id) != 0 {
		return ctx.SendStatus(fiber.StatusCreated)
	} else {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}
}

func (fc *FileController) downloadFile(ctx *fiber.Ctx) error {
	fileUniqueName := ctx.Params("fileUniqueName")

	sess, err := fc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		fc.log.Panic("Cannot get session", zap.Error(err))
	}

	userData := sess.Get("userData").(dtos.User)
	file := fc.storageRpo.GetOwnedFileByUniqueName(fileUniqueName, userData.Id)

	if file == nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	readFilePath := path.Join(userData.HomeDirectory, file.UniqueName)

	decryptionKey, err := base64.StdEncoding.DecodeString(userData.CryptKey)

	if err != nil {
		fc.log.Error("Cannot decode decryption key", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot download file from the server"})
	}

	fileContent := fc.fileSrv.DecryptAndReadFileContent(readFilePath, decryptionKey)

	return ctx.Send(fileContent)
}

func (fc *FileController) getUserFiles(ctx *fiber.Ctx) error {
	sess, err := fc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		fc.log.Panic("Cannot get session", zap.Error(err))
	}

	userData := sess.Get("userData").(dtos.User)

	files := fc.storageRpo.GetOwnedFiles(userData.Id)

	return ctx.JSON(files)
}

func (fc *FileController) deleteFile(ctx *fiber.Ctx) error {
	fileUniqueName := ctx.Params("fileUniqueName")

	sess, err := fc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		fc.log.Panic("Cannot get session", zap.Error(err))
	}

	userData := sess.Get("userData").(dtos.User)
	file := fc.storageRpo.GetOwnedFileByUniqueName(fileUniqueName, userData.Id)

	if file == nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	deleteFilePath := path.Join(userData.HomeDirectory, file.UniqueName)

	if fc.fileSrv.RemoveFileFromDisk(deleteFilePath) == false {
		fc.log.Error("Cannot delete file from disk", zap.String("FilePath", deleteFilePath))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot delete file from the server"})
	}

	if fc.storageRpo.DeleteFile(file.UniqueName) == false {
		fc.log.Error("Cannot delete file from database", zap.String("UniqueFileName", file.UniqueName))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot delete file from the server"})
	}

	return ctx.SendStatus(fiber.StatusOK)
}
