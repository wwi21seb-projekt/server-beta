package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type PushSubscriptionRepositoryInterface interface {
	CreatePushSubscription(pushSubscription *models.PushSubscription) error
	GetPushSubscriptionsByUsername(username string) ([]models.PushSubscription, error)
	DeletePushSubscriptionById(pushSubscriptionId string) error
}

type PushSubscriptionRepository struct {
	DB *gorm.DB
}

// NewPushSubscriptionRepository can be used as a constructor to create a PushSubscriptionRepository "object"
func NewPushSubscriptionRepository(db *gorm.DB) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{DB: db}
}

func (repo *PushSubscriptionRepository) CreatePushSubscription(pushSubscription *models.PushSubscription) error {
	err := repo.DB.Create(pushSubscription).Error
	return err
}

func (repo *PushSubscriptionRepository) GetPushSubscriptionsByUsername(username string) ([]models.PushSubscription, error) {
	var pushSubscriptions []models.PushSubscription
	err := repo.DB.
		Where("username_fk = ?", username).
		Find(&pushSubscriptions).Error
	return pushSubscriptions, err
}

func (repo *PushSubscriptionRepository) DeletePushSubscriptionById(pushSubscriptionId string) error {
	err := repo.DB.Where("id = ?", pushSubscriptionId).Delete(&models.PushSubscription{}).Error
	return err
}
