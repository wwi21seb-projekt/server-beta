package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type PostRepositoryInterface interface {
	CreatePost(post *models.Post) error
	FindPostsByUserID(userID string, offset int, limit int) ([]models.Post, error)
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

// FindPostsByUserID retrieves posts for a user sorted by userid
func (repo *PostRepository) FindPostsByUserID(userID string, offset int, limit int) ([]models.Post, error) {
	var posts []models.Post

	// Execute the query to fetch the posts
	err := repo.DB.Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC"). // sorted by creation date in descending order
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return posts, nil
}
