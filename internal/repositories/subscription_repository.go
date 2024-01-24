package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type SubscriptionRepositoryInterface interface {
	CreateSubscription(subscription *models.Subscription) error
	DeleteSubscription(subscriptionId string) error
	GetSubscriptionByUsernames(follower, following string) (*models.Subscription, error)
	GetSubscriptionById(subscriptionId string) (*models.Subscription, error)
	GetSubscriptionCountByUsername(username string) (int64, int64, error)
	GetFollowers(limit int, offset int, currentUsername string) ([]models.Subscription, int64, error)
	GetFollowings(limit int, offset int, currentUsername string) ([]models.Subscription, int64, error)
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

// GetSubscriptionCountByUsername gets the number of followers and followings for a user
func (repo *SubscriptionRepository) GetSubscriptionCountByUsername(username string) (int64, int64, error) {
	var followerCount, followingCount int64

	err := repo.DB.Model(&models.Subscription{}).
		Where("following = ?", username). // following and follower are reversed because we want to know who is following the user
		Count(&followerCount).
		Error
	if err != nil {
		return 0, 0, err
	}

	err = repo.DB.Model(&models.Subscription{}).Where("follower = ?", username).Count(&followingCount).Error
	if err != nil {
		return 0, 0, err
	}

	return followerCount, followingCount, nil
}

func (repo *SubscriptionRepository) GetFollowers(limit int, offset int, username string) ([]models.Subscription, int64, error) {
	var followers []models.Subscription
	var count int64

	baseQuery := repo.DB.Model(&models.Subscription{}).
		Where("following = ?", username)

	// Count results
	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Get users
	err = baseQuery.Limit(limit).Offset(offset).Preload("Following").Find(&followers).Error
	if err != nil {
		return nil, 0, err
	}

	return followers, count, nil
}
func (repo *SubscriptionRepository) GetFollowings(limit int, offset int, username string) ([]models.Subscription, int64, error) {
	var followings []models.Subscription
	var count int64

	baseQuery := repo.DB.Model(&models.Subscription{}).
		Where("follower = ?", username)

	// Count results
	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Get users
	err = baseQuery.Limit(limit).Offset(offset).Preload("Follower").Find(&followings).Error
	if err != nil {
		return nil, 0, err
	}

	return followings, count, nil
}
