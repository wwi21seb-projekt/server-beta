package repositories

import (
	"errors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	FindUserByUsername(username string) (*models.User, error)
	BeginTx() *gorm.DB
	CommitTx(tx *gorm.DB) error
	RollbackTx(tx *gorm.DB)
	CreateUserTx(user *models.User, tx *gorm.DB) error
	CheckEmailExistsForUpdate(email string, tx *gorm.DB) (bool, error)
	CheckUsernameExistsForUpdate(username string, tx *gorm.DB) (bool, error)
	UpdateUser(user *models.User) error
	SearchUser(username string, limit int, offset int, currentUsername string) ([]models.User, int64, error)
	GetUnactivatedUsers() ([]models.User, error)
	DeleteUserByUsername(username string) error
}

type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository can be used as a constructor to create a UserRepository "object"
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (repo *UserRepository) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := repo.DB.Where("username = ?", username).First(&user).Error
	return &user, err
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

func (repo *UserRepository) SearchUser(username string, limit int, offset int, currentUsername string) ([]models.User, int64, error) {
	var users []models.User
	var count int64
	maxLevenshteinDistance := 3.5 // max distance for search results is set to ensure that only relevant results are returned

	query := repo.DB.Model(&models.User{}).
		Where("username != ?", currentUsername). // exclude current user from search
		Select("*, levenshtein(username, ?) as distance", username).
		Where("levenshtein(username, ?) <= ? OR username like ?", username, maxLevenshteinDistance, username+"%").
		Order("distance ASC")

	// Count results
	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Get users
	err = query.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

func (repo *UserRepository) GetUnactivatedUsers() ([]models.User, error) {
	var users []models.User
	err := repo.DB.Where("activated = ?", false).Find(&users).Error
	return users, err
}

func (repo *UserRepository) DeleteUserByUsername(username string) error {
	return repo.DB.Transaction(func(tx *gorm.DB) error {
		// Token löschen
		if err := tx.Where("username = ?", username).Delete(&models.ActivationToken{}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err // Rückkehr bei einem Datenbankfehler, der kein RecordNotFound-Fehler ist
			}
		}
		// Nutzer löschen
		if err := tx.Where("username = ?", username).Delete(&models.User{}).Error; err != nil {
			return err // Rückkehr bei einem Fehler
		}

		return nil // Kein Fehler, erfolgreicher Abschluss der Transaktion
	})
}
