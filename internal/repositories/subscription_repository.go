package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type SubscriptionRepositoryInterface interface {
	CreateSubscription(subscription *models.Subscription) error
	DeleteSubscription(subscriptionId string) error
	GetSubscriptionByUsernames(follower, following string) (*models.Subscription, error)
	GetSubscriptionById(subscriptionId string) (*models.Subscription, error)
}

type SubscriptionRepository struct {
	DB *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{DB: db}
}

func (repo *SubscriptionRepository) CreateSubscription(subscription *models.Subscription) error {
	return repo.DB.Create(subscription).Error
}

func (repo *SubscriptionRepository) DeleteSubscription(subscriptionId string) error {
	return repo.DB.Delete(&models.Subscription{}, "id = ?", subscriptionId).Error
}

func (repo *SubscriptionRepository) GetSubscriptionByUsernames(follower, following string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := repo.DB.Where("follower = ? AND following = ?", follower, following).First(&subscription).Error
	return &subscription, err
}

func (repo *SubscriptionRepository) GetSubscriptionById(subscriptionId string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := repo.DB.Where("id = ?", subscriptionId).First(&subscription).Error
	return &subscription, err
}
