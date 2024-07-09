package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type ImageRepositoryInterface interface {
	GetImageById(id string) (*models.Image, error)
	DeleteImageById(id string) error
}

type ImageRepository struct {
	DB *gorm.DB
}

// NewImageRepository can be used as a constructor to create a ImageRepository "object"
func NewImageRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{DB: db}
}

func (repo *ImageRepository) GetImageById(id string) (*models.Image, error) {
	var image models.Image
	err := repo.DB.Where("id = ?", id).First(&image).Error
	return &image, err
}

func (repo *ImageRepository) DeleteImageById(id string) error {
	err := repo.DB.Where("id = ?", id).Delete(&models.Image{}).Error
	return err
}
