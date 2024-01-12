package repositories

import (
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type PostRepositoryInterface interface {
	CreatePost(post *models.Post) error
	GetPosts(lastPostId uuid.UUID, limit int) ([]models.Post, error)
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

func (repo *PostRepository) GetPosts(lastPostId uuid.UUID, limit int) ([]models.Post, error) {
	var posts []models.Post

	if lastPostId == uuid.Nil {
		err := repo.DB.Order("created_at desc").Limit(limit).Find(&posts).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := repo.DB.Where("id > ?", lastPostId).Order("created_at desc").Limit(limit).Find(&posts).Error
		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}
