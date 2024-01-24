package repositories

import (
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type PostRepositoryInterface interface {
	CreatePost(post *models.Post) error
	GetPostCountByUsername(username string) (int64, error)
	GetPostsByUsername(username string, offset, limit int) ([]models.Post, int64, error)
	GetPostById(postId string) (models.Post, error)
	GetPostsGlobalFeed(lastPost *models.Post, limit int) ([]models.Post, int64, error)
	GetPostsPersonalFeed(username string, lastPost *models.Post, limit int) ([]models.Post, int64, error)
	DeletePostById(postId string) error
	GetPostsByHashtag(hashtag string, lastPost *models.Post, limit int) ([]models.Post, int64, error)
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
	err := repo.DB.Where("id = ?", postId).First(&post).Error
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
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

func (repo *PostRepository) DeletePostById(postId string) error {
	return repo.DB.Where("id = ?", postId).Delete(&models.Post{}).Error
}

func (repo *PostRepository) GetPostsByHashtag(hashtag string, lastPost *models.Post, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var count int64
	var err error

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Disable color
		},
	)
	repo.DB.Logger = newLogger

	baseQuery := repo.DB.Model(&models.Post{}).
		Joins("JOIN post_hashtags ON post_hashtags.post_id = posts.id").
		Joins("JOIN hashtags ON hashtags.id = post_hashtags.hashtag_id").
		Where("hashtags.name = ?", hashtag)

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
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, err
}
