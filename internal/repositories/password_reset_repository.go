package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type PasswordResetRepositoryInterface interface {
	CreatePasswordResetToken(token *models.PasswordResetToken) error
	FindPasswordResetToken(username string, token string) (*models.PasswordResetToken, error)
	DeletePasswordResetTokenById(id string) error
	DeletePasswordResetTokensByUsername(username string) error
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

func (repo *PasswordResetRepository) FindPasswordResetToken(username string, token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	err := repo.DB.Where("username_fk = ? AND token = ?", username, token).First(&resetToken).Error
	return &resetToken, err
}

func (repo *PasswordResetRepository) DeletePasswordResetTokenById(id string) error {
	return repo.DB.Where("id = ?", id).Delete(&models.PasswordResetToken{}).Error
}

func (repo *PasswordResetRepository) DeletePasswordResetTokensByUsername(username string) error {
	return repo.DB.Where("username_fk = ?", username).Delete(&models.PasswordResetToken{}).Error
}
