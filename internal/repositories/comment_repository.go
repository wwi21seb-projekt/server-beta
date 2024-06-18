package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type CommentRepositoryInterface interface {
	CreateComment(comment *models.Comment) error
	GetCommentsByPostId(postId string, offset, limit int) ([]models.Comment, int64, error)
	CountComments(postId string) (int64, error)
}

type CommentRepository struct {
	DB *gorm.DB
}

// NewCommentRepository can be used as a constructor to create a CommentRepository "object"
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{DB: db}
}

func (repo *CommentRepository) CreateComment(comment *models.Comment) error {
	err := repo.DB.Create(comment).Error
	return err
}

func (repo *CommentRepository) GetCommentsByPostId(postId string, offset, limit int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var count int64

	baseQuery := repo.DB.Model(&models.Comment{}).Where("post_id = ?", postId)

	// Count number of comments based on post id
	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Get comments using pagination information
	err = baseQuery.
		Offset(offset).
		Limit(limit).
		Order("created_at desc, id desc").
		Preload("User").
		Preload("User.Image").
		Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, count, nil
}

func (repo *CommentRepository) CountComments(postId string) (int64, error) {
	var count int64
	query := repo.DB.Model(&models.Comment{}).Where("post_id = ?", postId)
	err := query.Count(&count).Error
	return count, err
}
