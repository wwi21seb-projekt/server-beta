package repositories

import (
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type LocationRepositoryInterface interface {
	CreateLocation(location *models.Location) error
	GetLocationById(Id uuid.UUID) (models.Location, error)
	DeleteLocationById(locationId string) error
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
func (repo *LocationRepository) GetLocationById(Id uuid.UUID) (models.Location, error) {
	var location models.Location
	err := repo.DB.Where("id = ?", Id).First(&location).Error
	return location, err
}

func (repo *LocationRepository) DeleteLocationById(locationId string) error {
	return repo.DB.Where("id = ?", locationId).Delete(&models.Location{}).Error
}
