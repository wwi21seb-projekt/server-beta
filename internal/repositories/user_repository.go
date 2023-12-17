package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	FindUserByUsername(username string) (models.User, error)
	BeginTx() *gorm.DB
	CommitTx(tx *gorm.DB) error
	RollbackTx(tx *gorm.DB)
	CreateUserTx(user *models.User, tx *gorm.DB) error
	CheckEmailExistsForUpdate(email string, tx *gorm.DB) (bool, error)
	CheckUsernameExistsForUpdate(username string, tx *gorm.DB) (bool, error)
	UpdateUser(user *models.User) error
}

type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository can be used as a constructor to create a UserRepository "object"
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (repo *UserRepository) FindUserByUsername(username string) (models.User, error) {
	var user models.User
	err := repo.DB.Where("username = ?", username).First(&user).Error
	return user, err
}

func (repo *UserRepository) BeginTx() *gorm.DB {
	return repo.DB.Begin()
}

func (repo *UserRepository) CommitTx(tx *gorm.DB) error {
	return tx.Commit().Error
}

func (repo *UserRepository) RollbackTx(tx *gorm.DB) {
	tx.Rollback()
}

func (repo *UserRepository) CreateUserTx(user *models.User, tx *gorm.DB) error {
	err := tx.Create(&user).Error
	return err
}

func (repo *UserRepository) CheckEmailExistsForUpdate(email string, tx *gorm.DB) (bool, error) {
	var count int64
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *UserRepository) CheckUsernameExistsForUpdate(username string, tx *gorm.DB) (bool, error) {
	var count int64
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *UserRepository) UpdateUser(user *models.User) error {
	err := repo.DB.Save(&user).Error
	return err
}
