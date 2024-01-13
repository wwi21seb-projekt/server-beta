package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/repositories"
	"github.com/marcbudd/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"time"
	"unicode/utf8"
)

type PostServiceInterface interface {
	CreatePost(req *models.PostCreateRequestDTO, username string) (*models.PostCreateResponseDTO, *customerrors.CustomError, int)
	FindPostsByUser(username string, offset, limit int) (*models.UserFeedDTO, *customerrors.CustomError, int)
}

type PostService struct {
	postRepo    repositories.PostRepositoryInterface
	userRepo    repositories.UserRepositoryInterface
	hashtagRepo repositories.HashtagRepositoryInterface
}

// NewPostService can be used as a constructor to create a PostService "object"
func NewPostService(postRepo repositories.PostRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	hashtagRepo repositories.HashtagRepositoryInterface) *PostService {
	return &PostService{postRepo: postRepo, userRepo: userRepo, hashtagRepo: hashtagRepo}
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

	//Extract hashtags
	hashtagNames := utils.ExtractHashtags(req.Content)

	// Create hashtag object for each hashtag name
	var hashtags []models.Hashtag
	for _, name := range hashtagNames {
		hashtag, err := service.hashtagRepo.FindOrCreateHashtag(name)
		if err != nil {
			return nil, customerrors.InternalServerError, http.StatusInternalServerError
		}
		hashtags = append(hashtags, hashtag)
	}

	// Create post
	post := models.Post{
		Id:        uuid.New(),
		Username:  username,
		Content:   req.Content,
		ImageUrl:  "",
		Hashtags:  hashtags,
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

func (service *PostService) FindPostsByUser(username string, offset, limit int) (*models.UserFeedDTO, *customerrors.CustomError, int) {

	// See if user exists
	_, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Get posts total count
	totalPostsCount, err := service.postRepo.FindPostsByUserCount(username)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	posts, err := service.postRepo.FindPostsByUser(username, offset, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto and return
	var postDtos []models.UserFeedRecordDTO
	for _, post := range posts {
		postDto := models.UserFeedRecordDTO{
			PostId:       post.Id,
			CreationDate: post.CreatedAt,
			Content:      post.Content,
		}
		postDtos = append(postDtos, postDto)
	}

	paginationDto := models.UserFeedPaginationDTO{
		Offset:  offset,
		Limit:   limit,
		Records: totalPostsCount,
	}

	userFeedDto := models.UserFeedDTO{
		Records:    postDtos,
		Pagination: &paginationDto,
	}

	return &userFeedDto, nil, http.StatusOK
}
