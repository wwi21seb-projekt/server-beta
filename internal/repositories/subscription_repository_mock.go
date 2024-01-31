package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) CreateSubscription(subscription *models.Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) DeleteSubscription(subscriptionId string) error {
	args := m.Called(subscriptionId)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetSubscriptionByUsernames(follower, following string) (*models.Subscription, error) {
	args := m.Called(follower, following)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetSubscriptionById(subscriptionId string) (*models.Subscription, error) {
	args := m.Called(subscriptionId)
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetSubscriptionCountByUsername(username string) (int64, int64, error) {
	args := m.Called(username)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockSubscriptionRepository) GetFollowers(limit int, offset int, username string, currentUsername string) ([]models.UserSubscriptionRecordDTO, int64, error) {
	args := m.Called(limit, offset, username, currentUsername)
	return args.Get(0).([]models.UserSubscriptionRecordDTO), args.Get(1).(int64), args.Error(2)
}

func (m *MockSubscriptionRepository) GetFollowings(limit int, offset int, username string, currentUsername string) ([]models.UserSubscriptionRecordDTO, int64, error) {
	args := m.Called(limit, offset, username, currentUsername)
	return args.Get(0).([]models.UserSubscriptionRecordDTO), args.Get(1).(int64), args.Error(2)
}
