package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
)

type LikeServiceInterface interface {
	PostLike(postId, currentUsername string) (*customerrors.CustomError, int)
	DeleteLike(postId, currentUsername string) (*customerrors.CustomError, int)
}

type LikeService struct {
	likeRepo repositories.LikeRepositoryInterface
	postRepo repositories.PostRepositoryInterface
}

// NewLikeService can be used as a constructor to create a LikeService "object"
func NewLikeService(
	likeRepo repositories.LikeRepositoryInterface,
	postRepo repositories.PostRepositoryInterface) *LikeService {
	return &LikeService{likeRepo: likeRepo, postRepo: postRepo}
}

// PostLike creates a like for a given post id and the current logged-in user
func (service *LikeService) PostLike(postId, currentUsername string) (*customerrors.CustomError, int) {

	// Check if post exists
	post, err := service.postRepo.GetPostById(postId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.PostNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if like already exists
	_, err = service.likeRepo.FindLike(postId, currentUsername)
	if err == nil {
		return customerrors.AlreadyLiked, http.StatusConflict
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create like
	newLike := models.Like{
		Id:       uuid.New(),
		PostId:   post.Id,
		Username: currentUsername,
	}

	err = service.likeRepo.CreateLike(&newLike)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Return
	return nil, http.StatusNoContent
}

// DeleteLike deletes a like for a given post id and the current logged-in user
func (service *LikeService) DeleteLike(postId string, currentUsername string) (*customerrors.CustomError, int) {
	// Check if post exists
	_, err := service.postRepo.GetPostById(postId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.PostNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Get like
	like, err := service.likeRepo.FindLike(postId, currentUsername)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.NotLiked, http.StatusConflict
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Delete like
	err = service.likeRepo.DeleteLike(like.Id.String())
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}
