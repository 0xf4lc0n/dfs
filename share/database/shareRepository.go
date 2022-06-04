package database

import (
	"dfs/share/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type ShareRepository struct {
	logger   *zap.Logger
	database *gorm.DB
}

func NewShareRepository(log *zap.Logger, db *gorm.DB) *ShareRepository {
	return &ShareRepository{logger: log, database: db}
}

func (sr *ShareRepository) CreateShareFileEntry(fileId uint, sharedForId uint, sharedById uint, expirationTime time.Time) bool {
	share := models.Share{
		FileId:         fileId,
		SharedForId:    sharedForId,
		SharedById:     sharedById,
		ExpirationTime: expirationTime,
	}

	if err := sr.database.Create(&share).Error; err != nil {
		sr.logger.Error("Cannot create ShareSpace", zap.Error(err))
		return false
	}

	return true
}

func (sr *ShareRepository) DeleteShareFileEntry(fileId uint, sharedForId uint, sharedById uint) bool {
	share := models.Share{FileId: fileId, SharedForId: sharedForId, SharedById: sharedById}

	if err := sr.database.Delete(&share).Error; err != nil {
		sr.logger.Error("Cannot delete shared file entry", zap.Error(err))
		return false
	}

	return true
}

func (sr *ShareRepository) GetSharedForFileEntry(fileId uint, sharedForId uint) *models.Share {
	share := models.Share{FileId: fileId, SharedForId: sharedForId}

	if err := sr.database.First(&share).Error; err != nil {
		sr.logger.Error("Cannot find shared file entry", zap.Error(err))
		return nil
	}

	return &share
}

func (sr *ShareRepository) GetSharedForUserFilesEntries(sharedForId uint) []models.Share {
	var sharedFiles []models.Share

	if err := sr.database.Where("shared_for_id = ?", sharedForId).Find(&sharedFiles).Error; err != nil {
		sr.logger.Error("Cannot find shared files entries", zap.Error(err))
		return nil
	}

	return sharedFiles
}

func (sr *ShareRepository) GetSharedEntriesByFileId(fileId uint) []models.Share {
	var sharedFiles []models.Share

	if err := sr.database.Where("file_id = ?", fileId).Find(&sharedFiles).Error; err != nil {
		sr.logger.Error("Cannot find shared files entries", zap.Error(err))
		return nil
	}

	return sharedFiles
}

func (sr *ShareRepository) GetSharedByUserFilesEntries(sharedById uint) []models.Share {
	var sharedFiles []models.Share

	if err := sr.database.Where("shared_by_id = ?", sharedById).Find(&sharedFiles).Error; err != nil {
		sr.logger.Error("Cannot find shared files entries", zap.Error(err))
		return nil
	}

	return sharedFiles
}
