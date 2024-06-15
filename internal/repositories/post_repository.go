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
	err := repo.DB.Create(post).Error
	return err
}

func (repo *PostRepository) GetPostCountByUsername(username string) (int64, error) {
	var count int64
	err := repo.DB.Model(&models.Post{}).Where("username_fk = ?", username).Count(&count).Error
	return count, err
}

func (repo *PostRepository) GetPostsByUsername(username string, offset, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var count int64

	baseQuery := repo.DB.Model(&models.Post{}).Where("username_fk = ?", username)

	// Count number of posts based on username
	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Get posts using pagination information
	err = baseQuery.
		Offset(offset).
		Limit(limit).
		Order("created_at desc, id desc").
		Preload("Image").
		Preload("Location").
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

func (repo *PostRepository) GetPostById(postId string) (models.Post, error) {
	var post models.Post
	err := repo.DB.Model(&models.Post{}).
		Preload("Location").
		Preload("Image").
		Preload("User").
		Preload("User.Image").
		Where("id = ?", postId).First(&post).Error
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
		Preload("Location").
		Preload("Image").
		Preload("User").
		Preload("User.Image").
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
		Joins("JOIN subscriptions ON subscriptions.following = posts.username_fk").
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
		Preload("Location").
		Preload("Image").
		Preload("User").
		Preload("User.Image").
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

func (repo *PostRepository) DeletePostById(postId string) error {
	return repo.DB.Transaction(func(tx *gorm.DB) error {

		var post models.Post
		result := tx.First(&post, "id = ?", postId)
		if result.Error != nil {
			return result.Error
		}

		// Delete comments
		if err := tx.Where("post_id = ?", post.Id).Delete(&models.Comment{}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		// Delete likes
		if err := tx.Where("post_id = ?", postId).Delete(&models.Like{}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		// Delete post
		if err := tx.Where("id = ?", postId).Delete(&models.Post{}).Error; err != nil {
			return err
		}

		// Delete hashtag associations
		if err := tx.Model(&models.Post{Id: post.Id}).Association("Hashtags").Clear(); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		// Delete location
		if err := tx.Where("id = ?", post.LocationId).Delete(&models.Location{}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		// Delete image
		if err := tx.Where("id = ?", post.ImageId).Delete(&models.Image{}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		return nil
	})
}

func (repo *PostRepository) GetPostsByHashtag(hashtag string, lastPost *models.Post, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var count int64
	var err error

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
		baseQuery = baseQuery.Where("(posts.created_at < ?) OR (posts.created_at = ? AND posts.id < ?)", lastPost.CreatedAt, lastPost.CreatedAt, lastPost.Id)
	}

	// Posts subset based on pagination
	err = baseQuery.
		Order("posts.created_at desc, posts.id desc").
		Limit(limit).
		Preload("Location").
		Preload("Image").
		Preload("User").
		Preload("User.Image").
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, err
}
