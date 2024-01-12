package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type PostRepositoryInterface interface {
	CreatePost(post *models.Post) error
	FindPostsByUser(username string, offset, limit int) ([]models.Post, error)
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

func (repo *PostRepository) FindPostsByUser(username string, offset, limit int) ([]models.Post, error) {
	var posts []models.Post
	err := repo.DB.Where("username = ?", username).Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}
