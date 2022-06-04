package database

import (
	"dfs/storage/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type StorageRepository struct {
	logger   *zap.Logger
	database *gorm.DB
}

func NewStorageRepository(logger *zap.Logger, database *gorm.DB) *StorageRepository {
	return &StorageRepository{logger: logger, database: database}
}

func (sr *StorageRepository) CreateFile(uniqueFileName string, fileName string, ownerId uint) uint {
	fileEntry := &models.File{
		UniqueName:   uniqueFileName,
		Name:         fileName,
		CreationDate: time.Now(),
		OwnerId:      ownerId,
	}

	if err := sr.database.Create(&fileEntry).Error; err != nil {
		sr.logger.Error("Cannot create new file entry", zap.Error(err))
		return 0
	}

	return fileEntry.Id
}

func (sr *StorageRepository) DeleteFile(uniqueFileName string) bool {
	fileEntry := models.File{UniqueName: uniqueFileName}

	if err := sr.database.Delete(&fileEntry).Error; err != nil {
		return false
	}

	return true
}

func (sr *StorageRepository) GetFileByName(uniqueFileName string) *models.File {
	var file models.File

	if err := sr.database.Where("unique_name = ?", uniqueFileName).First(&file).Error; err != nil {
		return nil
	}

	return &file
}

func (sr *StorageRepository) GetOwnedFileByName(uniqueFileName string, ownerId uint) *models.File {
	file := sr.GetFileByName(uniqueFileName)

	if file == nil {
		return nil
	}

	if file.OwnerId == ownerId {
		return file
	}

	return nil
}

func (sr *StorageRepository) GetOwnedFiles(ownerId uint) []models.File {
	var files []models.File

	sr.database.Where("owner_id = ?", ownerId).Find(&files)

	return files
}
