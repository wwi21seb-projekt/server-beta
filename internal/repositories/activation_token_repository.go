package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type ActivationTokenRepositoryInterface interface {
	FindTokenByUsername(username string) ([]models.ActivationToken, error)
	FindActivationToken(username, token string) (*models.ActivationToken, error)
	DeleteActivationTokenByUsername(username string) error
}

type ActivationTokenRepository struct {
	DB *gorm.DB
}

// NewActivationTokenRepository can be used as a constructor to create a ActivationTokenRepository "object"
func NewActivationTokenRepository(db *gorm.DB) *ActivationTokenRepository {
	return &ActivationTokenRepository{DB: db}
}

func (repo *ActivationTokenRepository) FindTokenByUsername(username string) ([]models.ActivationToken, error) {
	var tokens []models.ActivationToken
	err := repo.DB.Where("username = ?", username).Find(&tokens).Error
	return tokens, err
}

func (repo *ActivationTokenRepository) FindActivationToken(username string, token string) (*models.ActivationToken, error) {
	var activationToken models.ActivationToken
	err := repo.DB.Where("username = ? AND token = ?", username, token).First(&activationToken).Error
	if err != nil {
		return nil, err
	}
	return &activationToken, nil
}

func (repo *ActivationTokenRepository) DeleteActivationTokenByUsername(username string) error {
	err := repo.DB.Where("username = ?", username).Delete(models.ActivationToken{}).Error
	return err
}
