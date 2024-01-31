package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type LocationRepositoryInterface interface {
	CreateLocation(location *models.Location) error
}

type LocationRepository struct {
	DB *gorm.DB
}

// NewLocationRepository can be used as a constructor to create a LocationRepository "object"
func NewLocationRepository(db *gorm.DB) *LocationRepository {
	return &LocationRepository{DB: db}
}

func (repo *LocationRepository) CreateLocation(location *models.Location) error {
	return repo.DB.Create(&location).Error
}
