package database

import (
	"dfs/auth/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strings"
	"time"
)

type VerificationRepository struct {
	database *gorm.DB
	logger   *zap.Logger
}

func NewVerificationRepository(db *gorm.DB, log *zap.Logger) *VerificationRepository {
	return &VerificationRepository{database: db, logger: log}
}

func (vr *VerificationRepository) CreateAndReturnVerificationData(email string) *models.VerificationData {
	verificationData := models.VerificationData{
		Email:     email,
		ExpiresAt: time.Now().Add(time.Hour * 1),
		Code:      strings.Replace(uuid.New().String(), "-", "", -1),
	}

	if err := vr.database.Create(&verificationData).Error; err != nil {
		vr.logger.Error("Cannot create verification data entry", zap.Error(err))
		return nil
	}

	return &verificationData
}

func (vr *VerificationRepository) DeleteVerification(id uint) bool {
	verificationData := models.VerificationData{Id: id}

	if err := vr.database.Delete(&verificationData).Error; err != nil {
		return false
	}

	return true
}

func (vr *VerificationRepository) GetVerificationByCode(code string) *models.VerificationData {
	var verificationData models.VerificationData

	if err := vr.database.Where("code = ?", code).First(&verificationData).Error; err != nil {
		return nil
	}

	return &verificationData
}
