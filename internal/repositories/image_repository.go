package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type ImageRepositoryInterface interface {
	CreateImage(image *models.Image) error
	DeleteImage(imageUrl string) error
	GetImageMetadata(imageUrl string) (*models.Image, error)
}

type ImageRepository struct {
	DB *gorm.DB
}

func (repo *ImageRepository) CreateImage(image *models.Image) error {
	err := repo.DB.Create(image).Error
	return err
}

func (repo *ImageRepository) DeleteImage(imageUrl string) error {
	return repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("imageUrl = ?", imageUrl).Delete(&models.Post{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (repo *ImageRepository) GetImageMetadata(imageUrl string) (*models.Image, error) {
	var image models.Image
	err := repo.DB.Preload("Location").Preload("User").Preload("Image").Where("imageUrl = ?", imageUrl).First(&image).Error
	return &image, err
}
