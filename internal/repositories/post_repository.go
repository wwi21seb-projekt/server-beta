package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/gorm"
)

type PostRepositoryInterface interface {
	CreatePost(post *models.Post) error
	GetPostById(postId string) (models.Post, error)
	GetPostsGlobalFeed(lastPost *models.Post, limit int) ([]models.Post, error)
	GetPostsPersonalFeed(username string, lastPost *models.Post, limit int) ([]models.Post, error)
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

func (repo *PostRepository) GetPostById(postId string) (models.Post, error) {
	var post models.Post
	err := repo.DB.Where("id = ?", postId).First(&post).Error
	return post, err
}

func (repo *PostRepository) GetPostsGlobalFeed(lastPost *models.Post, limit int) ([]models.Post, error) {
	var posts []models.Post
	if lastPost == nil {
		err := repo.DB.Order("created_at desc, id desc").Limit(limit).Find(&posts).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := repo.DB.Where("(created_at < ?) OR (created_at = ? AND id < ?)", lastPost.CreatedAt, lastPost.CreatedAt, lastPost.Id).
			Order("created_at desc, id desc").Limit(limit).Find(&posts).Error
		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}

func (repo *PostRepository) GetPostsPersonalFeed(username string, lastPost *models.Post, limit int) ([]models.Post, error) {
	var posts []models.Post
	// TODO: change to use subscription based on username
	if lastPost == nil {
		err := repo.DB.Order("created_at desc, id desc").Limit(limit).Find(&posts).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := repo.DB.Where("(created_at < ?) OR (created_at = ? AND id < ?)", lastPost.CreatedAt, lastPost.CreatedAt, lastPost.Id).
			Order("created_at desc, id desc").Limit(limit).Find(&posts).Error
		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}
