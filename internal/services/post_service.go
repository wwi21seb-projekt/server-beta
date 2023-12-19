package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
	"time"
	"unicode/utf8"
)

type PostServiceInterface interface {
	CreatePost(req *models.PostCreateRequestDTO, username string) (*models.PostCreateResponseDTO, *customerrors.CustomError, int)
}

type PostService struct {
	postRepo repositories.PostRepositoryInterface
	userRepo repositories.UserRepositoryInterface
}

// NewPostService can be used as a constructor to create a PostService "object"
func NewPostService(postRepo repositories.PostRepositoryInterface, userRepo repositories.UserRepositoryInterface) *PostService {
	return &PostService{postRepo: postRepo, userRepo: userRepo}
}

func (service *PostService) CreatePost(req *models.PostCreateRequestDTO, username string) (*models.PostCreateResponseDTO, *customerrors.CustomError, int) {

	// Validations: 0-256 characters and utf8 characters
	if len(req.Content) <= 0 || len(req.Content) > 256 { // TODO: if image is present, content can be empty
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	if !utf8.ValidString(req.Content) {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Get user
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.PreliminaryUserUnauthorized, http.StatusUnauthorized // TODO: Custom error for unauthorized?
		}
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Create post
	post := models.Post{
		Id:        uuid.New(),
		Username:  username,
		User:      *user,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	err = service.postRepo.CreatePost(&post)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto and return
	authorDto := models.AuthorDTO{
		Username:          user.Username,
		Nickname:          user.Nickname,
		ProfilePictureUrl: user.ProfilePictureUrl,
	}
	postDto := models.PostCreateResponseDTO{
		PostId:       post.Id,
		Author:       &authorDto,
		CreationDate: post.CreatedAt,
		Content:      post.Content,
	}

	return &postDto, nil, http.StatusCreated
}
