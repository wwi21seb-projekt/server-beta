package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type PasswordResetRepositoryInterface interface {
	CreatePasswordResetToken(token *models.PasswordResetToken) error
	FindPasswordResetToken(token string) (*models.PasswordResetToken, error)
	DeletePasswordResetToken(token string) error
}

type PasswordResetRepository struct {
	DB *gorm.DB
}

// NewPasswordResetRepository can be used as a constructor to create a PasswordResetRepository "object"
func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{DB: db}
}

func (repo *PasswordResetRepository) CreatePasswordResetToken(token *models.PasswordResetToken) error {
	return repo.DB.Create(token).Error
}

func (repo *PasswordResetRepository) FindPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	err := repo.DB.Where("token = ?", token).First(&resetToken).Error
	return &resetToken, err
}

func (repo *PasswordResetRepository) DeletePasswordResetToken(token string) error {
	return repo.DB.Where("token = ?", token).Delete(&models.PasswordResetToken{}).Error
}
