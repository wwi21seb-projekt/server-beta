package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

// MockUserRepository is a mock implementation of the UserRepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) BeginTx() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockUserRepository) CommitTx(tx *gorm.DB) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockUserRepository) RollbackTx(tx *gorm.DB) {
	args := m.Called(tx)
	_ = args.Error(0)
}

func (m *MockUserRepository) CreateUserTx(user *models.User, tx *gorm.DB) error {
	args := m.Called(user, tx)
	return args.Error(0)
}

func (m *MockUserRepository) CheckEmailExistsForUpdate(email string, tx *gorm.DB) (bool, error) {
	args := m.Called(email, tx)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CheckUsernameExistsForUpdate(username string, tx *gorm.DB) (bool, error) {
	args := m.Called(username, tx)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) SearchUser(username string, limit int, offset int) ([]models.User, int64, error) {
	args := m.Called(username, limit, offset)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}
