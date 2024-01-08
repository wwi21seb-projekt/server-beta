package repositories

import (
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type SubscriptionRepositoryInterface interface {
	CreateSubscription(subscription *models.Subscription) error
	DeleteSubscription(subscriptionId uuid.UUID) error
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

func (repo *SubscriptionRepository) DeleteSubscription(subscriptionId uuid.UUID) error {
	// Finden und Löschen der Subscription
	return repo.DB.Delete(&models.Subscription{}, "subscription_id = ?", subscriptionId).Error
}
