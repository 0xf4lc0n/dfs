package database

import (
	"dfs/auth/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository struct {
	database *gorm.DB
	logger   *zap.Logger
}

func NewUserRepository(db *gorm.DB, log *zap.Logger) *UserRepository {
	return &UserRepository{database: db, logger: log}
}

func (ur *UserRepository) CreateUser(name string, email string, password []byte, key string) uint {
	user := models.User{
		Name:          name,
		Email:         email,
		Password:      password,
		Verified:      false,
		HomeDirectory: email,
		CryptKey:      key,
	}

	if err := ur.database.Create(&user).Error; err != nil {
		ur.logger.Error("Cannot create user", zap.Error(err))
		return 0
	}

	return user.Id
}

func (ur *UserRepository) DeleteUserByEmail(email string) bool {
	user := models.User{Email: email}

	if err := ur.database.Delete(&user).Error; err != nil {
		ur.logger.Error("Cannot delete user", zap.Error(err))
		return false
	}

	return true
}

func (ur *UserRepository) GetUserById(userId uint) *models.User {
	var user models.User

	if err := ur.database.Where("id = ?", userId).First(&user).Error; err != nil {
		return nil
	}

	return &user
}

func (ur *UserRepository) GetUserByEmail(email string) *models.User {
	var user models.User

	if err := ur.database.Where("email = ?", email).First(&user).Error; err != nil {
		return nil
	}

	return &user
}

func (ur *UserRepository) VerifyUser(user *models.User) bool {
	if err := ur.database.Model(&user).Update("verified", true).Error; err != nil {
		ur.logger.Error("Cannot update user", zap.Error(err))
		return false
	}

	return true
}
