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
	GetFollowers(limit int, offset int, username string, currentUsername string) ([]models.UserSubscriptionSQLRecordDTO, int64, error)
	GetFollowings(limit int, offset int, username string, currentUsername string) ([]models.UserSubscriptionSQLRecordDTO, int64, error)
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

// GetFollowers returns the followers of a user
func (repo *SubscriptionRepository) GetFollowers(limit int, offset int, username string, currentUsername string) ([]models.UserSubscriptionSQLRecordDTO, int64, error) {
	var followers []models.UserSubscriptionSQLRecordDTO
	var count int64

	baseQuery := repo.DB.Table("users").
		Select("userToCurrentUserSub.id as follower_id, currentUserToUserSub.id as following_id, users.username as username, users.nickname as nickname, images.id as image_id, images.format as format, images.width as width, images.height as height, images.tag as tag").
		Joins("INNER JOIN subscriptions on users.username = subscriptions.follower").                                                                                                        // join users table with subscriptions
		Joins("LEFT OUTER JOIN (SELECT * FROM subscriptions WHERE subscriptions.following = ?) AS userToCurrentUserSub ON users.username = userToCurrentUserSub.follower", currentUsername). // see if the user in the list is following the current user
		Joins("LEFT OUTER JOIN (SELECT * FROM subscriptions WHERE subscriptions.follower = ?) AS currentUserToUserSub ON users.username = currentUserToUserSub.following", currentUsername). // see if current user is following the user in the list
		Joins("LEFT OUTER JOIN images ON users.image_id = images.id").                                                                                                                       // Join with image table to get image attributes
		Where("subscriptions.following = ?", username)                                                                                                                                       // get only subscription where the user is followed

	// Count results
	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	//Get users
	err = baseQuery.Limit(limit).Offset(offset).Scan(&followers).Error
	if err != nil {
		return nil, 0, err
	}

	return followers, count, nil
}

// GetFollowings returns the users a given user is following
func (repo *SubscriptionRepository) GetFollowings(limit int, offset int, username string, currentUsername string) ([]models.UserSubscriptionSQLRecordDTO, int64, error) {
	var followings []models.UserSubscriptionSQLRecordDTO
	var count int64

	baseQuery := repo.DB.Table("users").
		Select("userToCurrentUserSub.id as follower_id, currentUserToUserSub.id as following_id, users.username as username, users.nickname as nickname, images.id as image_id, images.format as format, images.width as width, images.height as height, images.tag as tag").
		Joins("INNER JOIN subscriptions on users.username = subscriptions.following").                                                                                                       // join users table with subscriptions
		Joins("LEFT OUTER JOIN (SELECT * FROM subscriptions WHERE subscriptions.following = ?) AS userToCurrentUserSub ON users.username = userToCurrentUserSub.follower", currentUsername). // see if the user in the list is following the current user
		Joins("LEFT OUTER JOIN (SELECT * FROM subscriptions WHERE subscriptions.follower = ?) AS currentUserToUserSub ON users.username = currentUserToUserSub.following", currentUsername). // see if current user is following the user in the list
		Joins("LEFT OUTER JOIN images ON users.image_id = images.id").                                                                                                                       // Join with image table to get image attributes
		Where("subscriptions.follower = ?", username)                                                                                                                                        // get only subscription where the user follows someone

	// Count results
	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	//Get users
	err = baseQuery.Limit(limit).Offset(offset).Scan(&followings).Error
	if err != nil {
		return nil, 0, err
	}

	return followings, count, nil
}
