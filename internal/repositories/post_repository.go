package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type PostRepositoryInterface interface {
	CreatePost(post *models.Post) error
	GetPostCountByUsername(username string) (int64, error)
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

func (repo *PostRepository) GetPostCountByUsername(username string) (int64, error) {
	var count int64
	err := repo.DB.Model(&models.Post{}).
		Where("username = ? ", username).
		Count(&count).Error
	return count, err
}
