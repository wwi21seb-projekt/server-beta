package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type PostRepositoryInterface interface {
	CreatePost(post *models.Post) error
	FindPostsByUsername(username string, offset, limit int) ([]models.Post, int64, error)
}

type PostRepository struct {
	DB *gorm.DB
}

// NewPostRepository can be used as a constructor to create a PostRepository "object"
func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{DB: db}
}

func (repo *PostRepository) CreatePost(post *models.Post) error {
	return repo.DB.Create(&post).Error
}

func (repo *PostRepository) FindPostsByUsername(username string, offset, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var count int64

	baseQuery := repo.DB.Model(&models.Post{}).Where("username = ?", username)

	// Count number of posts based on username
	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Get posts using pagination information
	err = baseQuery.Offset(offset).Limit(limit).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}
