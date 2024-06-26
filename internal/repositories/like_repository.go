package repositories

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"gorm.io/gorm"
)

type LikeRepositoryInterface interface {
	CreateLike(like *models.Like) error
	DeleteLike(likeId string) error
	FindLike(postId string, currentUsername string) (*models.Like, error)
	CountLikes(postId string) (int64, error)
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

func (repo *LikeRepository) DeleteLike(likeId string) error {
	return repo.DB.Delete(&models.Like{}, "id = ?", likeId).Error
}

func (repo *LikeRepository) FindLike(postId string, currentUsername string) (*models.Like, error) {
	var like models.Like
	err := repo.DB.Where("post_id = ? AND username_fk = ?", postId, currentUsername).First(&like).Error
	return &like, err
}

func (repo *LikeRepository) CountLikes(postId string) (int64, error) {
	var count int64
	query := repo.DB.Model(&models.Like{}).Where("post_id = ?", postId)
	err := query.Count(&count).Error
	return count, err
}
