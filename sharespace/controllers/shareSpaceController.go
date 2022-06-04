package controllers

import (
	"crypto/rand"
	"dfs/sharespace/database"
	"dfs/sharespace/dtos"
	"dfs/sharespace/models"
	"dfs/sharespace/services"
	"encoding/base64"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"path"
	"strconv"
)

type ShareSpaceController struct {
	logger               *zap.Logger
	rpcClient            *services.RpcClient
	shareSpaceRepository *database.ShareSpaceRepository
	store                *session.Store
}

func NewShareSpaceController(logger *zap.Logger, rpcClient *services.RpcClient,
	shareSpaceRepository *database.ShareSpaceRepository, store *session.Store) *ShareSpaceController {

	return &ShareSpaceController{logger: logger, rpcClient: rpcClient, shareSpaceRepository: shareSpaceRepository, store: store}
}

func (ssc *ShareSpaceController) RegisterRoutes(app *fiber.App) {
	app.Post("/api/sharespace", ssc.createShareSpace)
	app.Delete("/api/sharespace/:shareSpaceId", ssc.deleteShareSpace)
	app.Get("/api/sharespace", ssc.getShareSpaces)
	app.Post("/api/sharespace/user", ssc.addToShareSpace)
	app.Delete("/api/sharespace/user", ssc.deleteFromShareSpace)
	app.Get("/api/sharespace/:shareSpaceId", ssc.getShareSpaceMembers)
	app.Post("/api/sharespace/:shareSpaceId/file", ssc.uploadFileToShareSpace)
	app.Delete("/api/sharespace/:shareSpaceId/file/:uniqueFileName", ssc.deleteFileFromShareSpace)
	app.Get("/api/sharespace/:shareSpaceId/file", ssc.getFilesFromShareSpace)
	app.Get("/api/sharespace/:shareSpaceId/file/:uniqueFileName", ssc.downloadFileFromShareSpace)
}

func (ssc *ShareSpaceController) createShareSpace(ctx *fiber.Ctx) error {
	createSsDto := new(dtos.CreateShareSpaceDto)

	if err := ctx.BodyParser(&createSsDto); err != nil {
		ssc.logger.Warn("Cannot parse CreateShareSpace data", zap.Error(err))
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	sess, err := ssc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		ssc.logger.Panic("Cannot get session", zap.Error(err))
	}

	createdBy := sess.Get("userData").(dtos.UserDto)

	homeDirPath := fmt.Sprintf("%s_%s", createdBy.Email, createSsDto.ShareSpaceName)

	if ssc.rpcClient.CreateHomeDirectory(homeDirPath) == false {
		ssc.logger.Warn("Cannot create home directory for ShareSpace")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot create ShareSpace"})
	}

	key := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		ssc.logger.Error("Cannot generate encryption key")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Cannot create ShareSpace",
		})
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)

	ssId := ssc.shareSpaceRepository.CreateShareSpace(createSsDto.ShareSpaceName, createdBy.Id, homeDirPath, encodedKey)

	if ssId == 0 {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot create ShareSpace"})
	}

	if ssc.shareSpaceRepository.AddUserToShareSpace(createdBy.Id, ssId, models.Owner); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot create ShareSpace"})
	}

	return ctx.SendStatus(fiber.StatusCreated)
}

func (ssc *ShareSpaceController) deleteShareSpace(ctx *fiber.Ctx) error {
	shareSpaceId, err := strconv.ParseUint(ctx.Params("shareSpaceId"), 10, 0)

	if err != nil {
		ssc.logger.Error("Cannot parse ShareSpaceId as uint", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	sess, err := ssc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		ssc.logger.Panic("Cannot get session", zap.Error(err))
	}

	deleteBy := sess.Get("userData").(dtos.UserDto)

	shareSpace := ssc.shareSpaceRepository.GetOwnedShareSpaceById(uint(shareSpaceId), deleteBy.Id)

	if shareSpace == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "cannot delete unowned ShareSpace"})
	}

	if ssc.shareSpaceRepository.DeleteEntireShareSpace(uint(shareSpaceId)) == false {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "cannot delete ShareSpace"})
	}

	ssc.rpcClient.DeleteFileFromDisk(dtos.DeleteFileDto{FilePath: shareSpace.HomeDirectory})

	return ctx.SendStatus(fiber.StatusOK)
}

func (ssc *ShareSpaceController) getShareSpaces(ctx *fiber.Ctx) error {
	sess, err := ssc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		ssc.logger.Panic("Cannot get session", zap.Error(err))
	}

	loggedUser := sess.Get("userData").(dtos.UserDto)

	shareSpaces := ssc.shareSpaceRepository.GetUserShareSpaces(loggedUser.Id)

	return ctx.Status(fiber.StatusOK).JSON(shareSpaces)
}

func (ssc *ShareSpaceController) addToShareSpace(ctx *fiber.Ctx) error {
	newMember := new(dtos.MemberDto)

	if err := ctx.BodyParser(&newMember); err != nil {
		ssc.logger.Error("Cannot parse AddUser data", zap.Error(err))
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	if ssc.shareSpaceRepository.AddUserToShareSpace(newMember.UserId, newMember.ShareSpaceId, models.Member) {
		return ctx.SendStatus(fiber.StatusOK)
	} else {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot add user to sharespace"})
	}
}

func (ssc *ShareSpaceController) deleteFromShareSpace(ctx *fiber.Ctx) error {
	var memberDto dtos.MemberDto

	if err := ctx.BodyParser(&memberDto); err != nil {
		ssc.logger.Error("Cannot parse DeleteUser data", zap.Error(err))
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	sess, err := ssc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		ssc.logger.Panic("Cannot get session", zap.Error(err))
	}

	deleteBy := sess.Get("userData").(dtos.UserDto)

	if ssc.shareSpaceRepository.CanUserDeleteMembers(deleteBy.Id, memberDto.ShareSpaceId) {
		if ssc.shareSpaceRepository.DeleteUserFromShareSpace(memberDto.UserId, memberDto.ShareSpaceId) {
			return ctx.SendStatus(fiber.StatusOK)
		}
	} else {
		ssc.logger.Error("User is not permitted for members deletion", zap.Uint("UserId", memberDto.UserId),
			zap.Uint("ShareSpaceId", memberDto.ShareSpaceId), zap.Error(err))
	}

	return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot delete user from ShareSpace"})
}

func (ssc *ShareSpaceController) getShareSpaceMembers(ctx *fiber.Ctx) error {
	shareSpaceId, err := strconv.ParseUint(ctx.Params("shareSpaceId"), 10, 0)

	if err != nil {
		ssc.logger.Error("Cannot parse ShareSpaceId as uint", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	users := ssc.shareSpaceRepository.GetShareSpaceMembers(uint(shareSpaceId))

	return ctx.JSON(users)
}

func (ssc *ShareSpaceController) uploadFileToShareSpace(ctx *fiber.Ctx) error {
	shareSpaceId, err := strconv.ParseUint(ctx.Params("shareSpaceId"), 10, 0)

	if err != nil {
		ssc.logger.Error("Cannot parse ShareSpaceId as uint", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	fileHeader, err := ctx.FormFile("file")

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	sess, err := ssc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		ssc.logger.Panic("Cannot get session", zap.Error(err))
	}

	ssc.logger.Debug("Uploading file to the ShareSpace", zap.Uint64("ShareSpaceId", shareSpaceId),
		zap.String("FileName", fileHeader.Filename))

	file, err := fileHeader.Open()

	if err != nil {
		ssc.logger.Error("Cannot open file", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	defer file.Close()

	fileContent, err := ioutil.ReadAll(file)

	if err != nil {
		ssc.logger.Error("Cannot read file", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	shareSpace := ssc.shareSpaceRepository.GetShareSpaceById(uint(shareSpaceId))

	if shareSpace == nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	uniqueFileName := uuid.New().String()
	savePath := path.Join(shareSpace.HomeDirectory, uniqueFileName)

	ssc.logger.Debug("File will be saved into", zap.String("ReadPath", savePath))

	encryptionKey, err := base64.StdEncoding.DecodeString(shareSpace.CryptKey)

	if err != nil {
		ssc.logger.Error("Cannot decode encryption key", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	saveFileDto := dtos.SaveFileDto{SavePath: savePath, Content: fileContent, EncryptionKey: encryptionKey}

	if ssc.rpcClient.SaveFileOnDisk(saveFileDto) == false {
		ssc.logger.Error("Cannot save file on the disk")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	fileSendBy := sess.Get("userData").(dtos.UserDto)

	if ssc.shareSpaceRepository.AddFileToShareSpace(uint(shareSpaceId), fileHeader.Filename, savePath, fileSendBy.Id) == 0 {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	return ctx.SendStatus(fiber.StatusCreated)
}

func (ssc *ShareSpaceController) deleteFileFromShareSpace(ctx *fiber.Ctx) error {
	shareSpaceId, err := strconv.ParseUint(ctx.Params("shareSpaceId"), 10, 0)

	if err != nil {
		ssc.logger.Error("Cannot parse ShareSpaceId as uint", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	uniqueFileName := ctx.Params("uniqueFileName")

	sess, err := ssc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		ssc.logger.Panic("Cannot get session", zap.Error(err))
	}

	fileDeletedBy := sess.Get("userData").(dtos.UserDto)

	fileToDelete := ssc.shareSpaceRepository.GetFileFromShareSpace(uint(shareSpaceId), uniqueFileName)

	if fileToDelete == nil {
		ssc.logger.Error("Cannot find file in the ShareSpace", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot delete unowned file"})
	}

	if fileToDelete.OwnerId != fileDeletedBy.Id {
		shareSpaceMember := ssc.shareSpaceRepository.GetShareSpaceMember(fileDeletedBy.Id, uint(shareSpaceId))

		if shareSpaceMember == nil {
			ssc.logger.Error("Cannot find user in the ShareSpace", zap.Error(err))
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot delete unowned file"})
		}

		if shareSpaceMember.Role == models.Member {
			ssc.logger.Error("User is not permitted to delete a file from ShareSpace",
				zap.Uint("UserId", fileDeletedBy.Id), zap.Uint("ShareSpaceId", uint(shareSpaceId)))
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot delete file "})
		}
	}

	deleteFileDto := dtos.DeleteFileDto{FilePath: fileToDelete.Path}

	if ssc.rpcClient.DeleteFileFromDisk(deleteFileDto) == false {
		ssc.logger.Error("Cannot delete file from disk")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot delete file"})
	}

	if ssc.shareSpaceRepository.DeleteFileFromShareSpace(uint(shareSpaceId), uniqueFileName) == false {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot delete unowned file"})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (ssc *ShareSpaceController) getFilesFromShareSpace(ctx *fiber.Ctx) error {
	shareSpaceId, err := strconv.ParseUint(ctx.Params("shareSpaceId"), 10, 0)

	if err != nil {
		ssc.logger.Error("Cannot parse ShareSpaceId as uint", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	shareSpaceFiles := ssc.shareSpaceRepository.GetFilesFromShareSpace(uint(shareSpaceId))

	return ctx.JSON(shareSpaceFiles)
}

func (ssc *ShareSpaceController) downloadFileFromShareSpace(ctx *fiber.Ctx) error {
	shareSpaceId, err := strconv.ParseUint(ctx.Params("shareSpaceId"), 10, 0)

	if err != nil {
		ssc.logger.Error("Cannot parse ShareSpaceId as uint", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot upload file to the server"})
	}

	sess, err := ssc.store.Get(ctx)
	defer sess.Destroy()

	if err != nil {
		ssc.logger.Panic("Cannot get session", zap.Error(err))
	}

	fileDownloadedBy := sess.Get("userData").(dtos.UserDto)

	if ssc.shareSpaceRepository.IsUserMemberOfShareSpace(fileDownloadedBy.Id, uint(shareSpaceId)) == false {
		ssc.logger.Error("User isn't member of ShareSpace",
			zap.Uint64("ShareSpaceId", shareSpaceId), zap.Uint("UserId", fileDownloadedBy.Id))
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "user isn't member of ShareSpace"})
	}

	uniqueFileName := ctx.Params("uniqueFileName")

	shareSpaceFile := ssc.shareSpaceRepository.GetFileFromShareSpace(uint(shareSpaceId), uniqueFileName)

	if shareSpaceFile == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "file doesn't exist in the ShareSpace"})
	}

	shareSpace := ssc.shareSpaceRepository.GetShareSpaceById(uint(shareSpaceId))

	if shareSpace == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "file doesn't exist in the ShareSpace"})
	}

	decryptionKey, err := base64.StdEncoding.DecodeString(shareSpace.CryptKey)

	if err != nil {
		ssc.logger.Error("Cannot decode encryption key", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "cannot download file from the server"})
	}

	readFileDto := dtos.ReadFileDto{ReadPath: shareSpaceFile.Path, DecryptionKey: decryptionKey}

	fileContent := ssc.rpcClient.ReadFileFromDisk(readFileDto)

	return ctx.Status(fiber.StatusOK).Send(fileContent)
}
