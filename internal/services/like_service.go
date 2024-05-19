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
	PostLike(req *models.LikePostRequestDTO, currentUsername string) (*models.LikePostResponseDTO, *customerrors.CustomError, int)
	DeleteLike(likeId string, currentUsername string) (*customerrors.CustomError, int)
}

type LikeService struct {
	likeRepo repositories.LikeRepositoryInterface
	userRepo repositories.UserRepositoryInterface
	postRepo repositories.PostRepositoryInterface
}

func NewLikeService(
	likeRepo repositories.LikeRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	postRepo repositories.PostRepositoryInterface) *LikeService {
	return &LikeService{likeRepo: likeRepo, userRepo: userRepo, postRepo: postRepo}
}

func (service *LikeService) PostLike(req *models.LikePostRequestDTO, currentUsername string) (*models.LikePostResponseDTO, *customerrors.CustomError, int) {

	// Check if post exists
	_, err := service.postRepo.GetPostById(req.LikedPostId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.PostNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if like already exists
	_, err = service.likeRepo.FindLike(req.LikedPostId, currentUsername)
	if err == nil {
		return nil, customerrors.LikeAlreadyExists, http.StatusConflict
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create like
	newLike := models.Like{
		Id:           uuid.New(),
		LikedPostId:  req.LikedPostId,
		LikeUsername: currentUsername,
	}

	err = service.likeRepo.CreateLike(&newLike)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	//Create response
	response := &models.LikePostResponseDTO{}

	// Return
	return response, nil, http.StatusNoContent
}

func (service *LikeService) DeleteLike(postId string, currentUsername string) (*customerrors.CustomError, int) {

	// Get like
	like, err := service.likeRepo.FindLike(postId, currentUsername)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.LikeNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Delete like
	err = service.likeRepo.DeleteLike(like.Id)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}
