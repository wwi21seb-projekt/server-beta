package repositories

import (
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type PostRepositoryInterface interface {
	CreatePost(post *models.Post) error
	GetPostCountByUsername(username string) (int64, error)
	GetPostsByUsername(username string, offset, limit int) ([]models.Post, int64, error)
	GetPostById(postId string) (models.Post, error)
	GetPostsGlobalFeed(lastPost *models.Post, limit int) ([]models.Post, int64, error)
	GetPostsPersonalFeed(username string, lastPost *models.Post, limit int) ([]models.Post, int64, error)
	DeletePostById(postId string) error
}

type PostRepository struct {
	DB *gorm.DB
}

// NewPostRepository can be used as a constructor to create a PostRepository "object"
func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{DB: db}
}

func (repo *PostRepository) CreatePost(post *models.Post) error {

	err := repo.DB.Transaction(func(tx *gorm.DB) error {
		// Erstellen des Posts
		if err := tx.Create(post).Error; err != nil {
			return err // Rückkehr bei einem Fehler
		}

		// Überprüfen, ob eine Location vorhanden ist und sie erstellen, wenn ja
		if post.Location != (models.Location{}) {
			// Die Location-ID mit der Post-ID verknüpfen
			post.Location.Id = post.Id
			if err := tx.Create(&post.Location).Error; err != nil {
				return err // Rückkehr bei einem Fehler
			}

		}

		return nil // Kein Fehler, erfolgreicher Abschluss der Transaktion
	})

	return err
}

func (repo *PostRepository) GetPostCountByUsername(username string) (int64, error) {
	var count int64
	err := repo.DB.Model(&models.Post{}).Where("username = ?", username).Count(&count).Error
	return count, err
}

func (repo *PostRepository) GetPostsByUsername(username string, offset, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var count int64

	baseQuery := repo.DB.Model(&models.Post{}).Where("username = ?", username)

	// Count number of posts based on username
	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Get posts using pagination information
	err = baseQuery.Offset(offset).Limit(limit).Order("created_at desc, id desc").Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

func (repo *PostRepository) GetPostById(postId string) (models.Post, error) {
	var post models.Post
	err := repo.DB.Preload("Location").Where("id = ?", postId).First(&post).Error
	return post, err
}

func (repo *PostRepository) GetPostsGlobalFeed(lastPost *models.Post, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var count int64
	var err error

	baseQuery := repo.DB.Model(&models.Post{})

	// Number of posts in global feed
	err = baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	if lastPost.Id != uuid.Nil {
		baseQuery = baseQuery.Where("(created_at < ?) OR (created_at = ? AND id < ?)", lastPost.CreatedAt, lastPost.CreatedAt, lastPost.Id)
	}

	// Posts subset based on pagination
	err = baseQuery.
		Order("created_at desc, id desc").
		Limit(limit).
		Preload("User").
		Preload("Location").
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, err
}

func (repo *PostRepository) GetPostsPersonalFeed(username string, lastPost *models.Post, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var count int64
	var err error

	baseQuery := repo.DB.Model(&models.Post{}).
		Joins("JOIN subscriptions ON subscriptions.following = posts.username").
		Where("subscriptions.follower = ?", username)

	// Number of posts in global feed
	err = baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	if lastPost.Id != uuid.Nil {
		baseQuery = baseQuery.Where("(created_at < ?) OR (created_at = ? AND posts.id < ?)", lastPost.CreatedAt, lastPost.CreatedAt, lastPost.Id)
	}

	// Posts subset based on pagination
	err = baseQuery.Order("created_at desc, posts.id desc").
		Limit(limit).
		Preload("User").
		Preload("Location").
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

func (repo *PostRepository) DeletePostById(postId string) error {
	return repo.DB.Transaction(func(tx *gorm.DB) error {

		var post models.Post
		result := repo.DB.First(&post, postId)
		if result.Error != nil {
			return result.Error
		}

		// Löschen der Hashtags-Beziehungen in der Join-Tabelle
		if err := tx.Model(&models.Post{Id: post.Id}).Association("Hashtags").Clear(); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err // Rückkehr bei einem Datenbankfehler, der kein RecordNotFound-Fehler ist
			}
		}
		// Löschen der Location
		if err := tx.Where("location_id = ?", post.LocationId).Delete(&models.Location{}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err // Rückkehr bei einem Datenbankfehler, der kein RecordNotFound-Fehler ist
			}
		}
		// Löschen des Posts
		if err := tx.Where("id = ?", postId).Delete(&models.Post{}).Error; err != nil {
			return err // Rückkehr bei einem Fehler
		}

		return nil // Kein Fehler, erfolgreicher Abschluss der Transaktion
	})
}
