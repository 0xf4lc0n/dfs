package database

import (
	"dfs/sharespace/dtos"
	"dfs/sharespace/models"
	"dfs/sharespace/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type ShareSpaceRepository struct {
	logger    *zap.Logger
	database  *gorm.DB
	rpcClient *services.RpcClient
}

func NewShareSpaceRepository(logger *zap.Logger, database *gorm.DB, rpcClient *services.RpcClient) *ShareSpaceRepository {
	return &ShareSpaceRepository{logger: logger, database: database, rpcClient: rpcClient}
}

func (ssr *ShareSpaceRepository) CreateShareSpace(shareSpaceName string, ownerId uint, homeDir string, key string) uint {
	shareSpace := models.ShareSpace{
		Name:          shareSpaceName,
		Owner:         ownerId,
		HomeDirectory: homeDir,
		CryptKey:      key,
	}

	if err := ssr.database.Create(&shareSpace).Error; err != nil {
		ssr.logger.Error("Cannot create ShareSpace", zap.Error(err))
		return 0
	}

	return shareSpace.Id
}

func (ssr *ShareSpaceRepository) DeleteShareSpace(ssId uint) bool {
	shareSpace := models.ShareSpace{Id: ssId}

	if err := ssr.database.Delete(&shareSpace).Error; err != nil {
		return false
	}

	return true
}

func (ssr *ShareSpaceRepository) GetShareSpaceById(ssId uint) *models.ShareSpace {
	shareSpace := models.ShareSpace{Id: ssId}

	if err := ssr.database.First(&shareSpace).Error; err != nil {
		ssr.logger.Error("Cannot find ShareSpace with the given ID", zap.Error(err))
		return nil
	}

	return &shareSpace
}

func (ssr *ShareSpaceRepository) GetUserShareSpaces(userId uint) []models.ShareSpace {
	var memberIn []models.ShareSpaceMember

	ssr.database.Where("user_id = ?", userId).Find(&memberIn)

	var shareSpaces []models.ShareSpace

	for _, mi := range memberIn {
		var shareSpace models.ShareSpace
		ssr.database.Where("id = ?", mi.ShareSpaceId).First(&shareSpace)

		shareSpaces = append(shareSpaces, shareSpace)
	}

	return shareSpaces
}

func (ssr *ShareSpaceRepository) AddUserToShareSpace(userId uint, ssId uint, role models.Role) bool {
	ssMember := models.ShareSpaceMember{
		ShareSpaceId: ssId,
		UserId:       userId,
		Role:         role,
	}

	if err := ssr.database.Create(&ssMember).Error; err != nil {
		ssr.logger.Error("Cannot add user to ShareSpace", zap.Uint("UserId", userId),
			zap.Uint("ShareSpaceId", ssId), zap.Error(err))
		return false
	}

	return true
}

func (ssr *ShareSpaceRepository) DeleteUserFromShareSpace(userId uint, ssId uint) bool {
	shareSpaceMember := models.ShareSpaceMember{
		ShareSpaceId: ssId,
		UserId:       userId,
	}

	var shareSpace models.ShareSpace

	if err := ssr.database.Where("id = ?", ssId).First(&shareSpace).Error; err != nil {
		ssr.logger.Error("Cannot get ShareSpace with given id", zap.Uint("ShareSpaceId", ssId), zap.Error(err))
		return false
	}

	if err := ssr.database.Delete(&shareSpaceMember).Error; err != nil {
		ssr.logger.Error("Cannot delete user form ShareSpace", zap.Uint("ShareSpaceId", ssId),
			zap.Uint("UserId", userId), zap.Error(err))
		return false
	}

	return true
}

func (ssr *ShareSpaceRepository) GetShareSpaceMember(userId uint, ssId uint) *models.ShareSpaceMember {
	var ssMember models.ShareSpaceMember

	if err := ssr.database.Where("user_id = ? AND share_space_id = ?", userId, ssId).First(&ssMember).Error; err != nil {
		ssr.logger.Error("Cannot get ShareSpace member", zap.Uint("UserId", userId),
			zap.Uint("ShareSpaceId", ssId), zap.Error(err))
		return nil
	}

	return &ssMember
}

func (ssr *ShareSpaceRepository) GetShareSpaceMembers(ssId uint) []dtos.UserDto {
	var ssMembers []models.ShareSpaceMember
	var users []dtos.UserDto

	ssr.database.Where("share_space_id = ?", ssId).Find(&ssMembers)

	for _, member := range ssMembers {
		user := ssr.rpcClient.GetUserDataById(member.UserId)
		users = append(users, *user)
	}

	return users
}

func (ssr *ShareSpaceRepository) IsUserMemberOfShareSpace(userId uint, ssId uint) bool {
	return ssr.GetShareSpaceMember(userId, ssId) != nil
}

func (ssr *ShareSpaceRepository) IsUserOwnerOfShareSpace(userId uint, ssId uint) bool {
	ssMember := ssr.GetShareSpaceMember(userId, ssId)

	if ssMember != nil {
		return ssMember.Role == models.Owner
	}

	return false
}

func (ssr *ShareSpaceRepository) CanUserDeleteMembers(userId uint, ssId uint) bool {
	ssMember := ssr.GetShareSpaceMember(userId, ssId)

	if ssMember != nil {
		return ssMember.Role == models.Owner || ssMember.Role == models.Moderator
	}

	return false
}

func (ssr *ShareSpaceRepository) AddFileToShareSpace(ssId uint, fileName string, savePath string, ownerId uint) uint {
	shareSpaceFileEntry := models.ShareSpaceFile{
		ShareSpaceId: ssId,
		Name:         fileName,
		UniqueName:   uuid.New().String(),
		Path:         savePath,
		CreationDate: time.Now(),
		OwnerId:      ownerId,
	}

	if err := ssr.database.Create(&shareSpaceFileEntry).Error; err != nil {
		ssr.logger.Error("Cannot create file entry in ShareSpace", zap.Error(err))
		return 0
	}

	return shareSpaceFileEntry.Id
}

func (ssr *ShareSpaceRepository) DeleteFileFromShareSpace(ssId uint, uniqueFileName string) bool {
	file := models.ShareSpaceFile{
		ShareSpaceId: ssId,
		UniqueName:   uniqueFileName,
	}

	if err := ssr.database.Delete(&file).Error; err != nil {
		ssr.logger.Error("Cannot delete file from database")
		return false
	}

	return true
}

func (ssr *ShareSpaceRepository) GetFileFromShareSpace(ssId uint, uniqueFileName string) *models.ShareSpaceFile {
	var file models.ShareSpaceFile

	if err := ssr.database.Where("share_space_id = ? AND unique_name = ?", ssId, uniqueFileName).
		First(&file).Error; err != nil {
		ssr.logger.Error("Cannot find file in the ShareSpace", zap.Error(err))
		return nil
	}

	return &file
}

func (ssr *ShareSpaceRepository) GetFilesFromShareSpace(ssId uint) []models.ShareSpaceFile {
	var shareSpaceFiles []models.ShareSpaceFile

	ssr.database.Where("share_space_id = ?", ssId).Find(&shareSpaceFiles)

	return shareSpaceFiles
}
