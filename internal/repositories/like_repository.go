package repositories

import (
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type LikeRepositoryInterface interface {
	CreateLike(like *models.Like) error
	DeleteLike(likeId uuid.UUID) error
	FindLike(postId string, currentUsername string) (*models.Like, error)
	CountLikes(postId string) int64
}

type LikeRepository struct {
	DB *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{DB: db}
}

func (repo *LikeRepository) CreateLike(like *models.Like) error {
	return repo.DB.Create(like).Error
}

func (repo *LikeRepository) DeleteLike(likeId uuid.UUID) error {
	return repo.DB.Delete(&models.Like{}, "id = ?", likeId).Error
}

func (repo *LikeRepository) FindLike(postId string, currentUsername string) (*models.Like, error) {
	var like models.Like
	err := repo.DB.Where("LikedPostId = ? AND LikeUser = ?", postId, currentUsername).First(&like).Error
	return &like, err
}

func (repo *LikeRepository) CountLikes(postId string) int64 {
	var count int64
	query := repo.DB.Where("LikedPostId = ?", postId)
	query.Count(&count)
	return count
}
