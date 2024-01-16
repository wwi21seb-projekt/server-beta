package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/repositories"
	"github.com/marcbudd/server-beta/internal/utils"
	"gorm.io/gorm"
	"mime/multipart"
	"net/http"
	"time"
	"unicode/utf8"
)

type PostServiceInterface interface {
	CreatePost(req *models.PostCreateRequestDTO, file *multipart.FileHeader, username string) (*models.PostCreateResponseDTO, *customerrors.CustomError, int)
	GetPostsByUsername(username string, offset, limit int) (*models.UserFeedDTO, *customerrors.CustomError, int)
	GetPostsGlobalFeed(lastPostId string, limit int) (*models.GeneralFeedDTO, *customerrors.CustomError, int)
	GetPostsPersonalFeed(username string, lastPostId string, limit int) (*models.GeneralFeedDTO, *customerrors.CustomError, int)
}

type PostService struct {
	postRepo     repositories.PostRepositoryInterface
	userRepo     repositories.UserRepositoryInterface
	hashtagRepo  repositories.HashtagRepositoryInterface
	imageService ImageServiceInterface
}

// NewPostService can be used as a constructor to create a PostService "object"
func NewPostService(postRepo repositories.PostRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	hashtagRepo repositories.HashtagRepositoryInterface,
	imageService ImageServiceInterface) *PostService {
	return &PostService{postRepo: postRepo, userRepo: userRepo, hashtagRepo: hashtagRepo, imageService: imageService}
}

func (service *PostService) CreatePost(req *models.PostCreateRequestDTO, file *multipart.FileHeader, username string) (*models.PostCreateResponseDTO, *customerrors.CustomError, int) {

	// Validations: 0-256 characters and utf8 characters
	if len(req.Content) > 256 {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	if len(req.Content) <= 0 && file == nil { // image or file can be empty, but not both
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	if !utf8.ValidString(req.Content) {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Get user
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserUnauthorized, http.StatusUnauthorized // TODO: Custom error for unauthorized?
		}
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Check if user is activated
	if !user.Activated {
		return nil, customerrors.PreliminaryUserUnauthorized, http.StatusUnauthorized
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

	// Save image if present
	var imageUrl = ""
	if file != nil {
		url, err, httpStatus := service.imageService.SaveImage(*file)
		if err != nil {
			return nil, err, httpStatus
		}
		imageUrl = url
	}

	// Create post
	post := models.Post{
		Id:        uuid.New(),
		Username:  username,
		Content:   req.Content,
		ImageUrl:  imageUrl,
		Hashtags:  hashtags,
		CreatedAt: time.Now(),
	}
	err = service.postRepo.CreatePost(&post)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto and return
	postDto := models.PostCreateResponseDTO{
		PostId: post.Id,
		Author: &models.AuthorDTO{
			Username:          user.Username,
			Nickname:          user.Nickname,
			ProfilePictureUrl: user.ProfilePictureUrl,
		},
		CreationDate: post.CreatedAt,
		Content:      post.Content,
	}

	return &postDto, nil, http.StatusCreated
}

func (service *PostService) GetPostsByUsername(username string, offset, limit int) (*models.UserFeedDTO, *customerrors.CustomError, int) {

	// See if user exists
	_, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Get posts
	posts, totalPostsCount, err := service.postRepo.GetPostsByUsername(username, offset, limit)
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

// GetPostsGlobalFeed returns a pagination object with the posts in the global feed using pagination parameters
func (service *PostService) GetPostsGlobalFeed(lastPostId string, limit int) (*models.GeneralFeedDTO, *customerrors.CustomError, int) {
	// Initialise empty GeneralFeedDTO
	feed := models.GeneralFeedDTO{
		Records:    []models.PostCreateResponseDTO{},
		Pagination: &models.GeneralFeedPaginationDTO{},
	}

	// Get last post if lastPostId is not empty
	var lastPost models.Post
	if lastPostId != "" {
		post, err := service.postRepo.GetPostById(lastPostId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, customerrors.PreliminaryPostNotFound, http.StatusNotFound
			}
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
		lastPost = post
	}

	// Retrieve posts from the database
	posts, totalPostsCount, err := service.postRepo.GetPostsGlobalFeed(&lastPost, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Fill GeneralFeedDTO with posts
	for _, post := range posts {
		authorDto := models.AuthorDTO{
			Username:          post.User.Username,
			Nickname:          post.User.Nickname,
			ProfilePictureUrl: post.User.ProfilePictureUrl,
		}
		postDto := models.PostCreateResponseDTO{
			PostId:       post.Id,
			Author:       &authorDto,
			CreationDate: post.CreatedAt,
			Content:      post.Content,
		}
		feed.Records = append(feed.Records, postDto)
	}

	// Set pagination details
	if len(posts) > 0 {
		feed.Pagination.LastPostId = posts[len(posts)-1].Id
	}
	feed.Pagination.Limit = limit
	feed.Pagination.Records = totalPostsCount

	return &feed, nil, http.StatusOK
}

// GetPostsPersonalFeed returns a pagination object with the posts in the personal feed using pagination parameters
func (service *PostService) GetPostsPersonalFeed(username string, lastPostId string, limit int) (*models.GeneralFeedDTO, *customerrors.CustomError, int) {
	// Initialise empty GeneralFeedDTO
	feed := models.GeneralFeedDTO{
		Records:    []models.PostCreateResponseDTO{},
		Pagination: &models.GeneralFeedPaginationDTO{},
	}

	// Get last post if lastPostId is not empty
	var lastPost models.Post
	if lastPostId != "" {
		post, err := service.postRepo.GetPostById(lastPostId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, customerrors.PreliminaryPostNotFound, http.StatusNotFound
			}
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
		lastPost = post
	}

	// Retrieve posts from the database
	posts, totalPostsCount, err := service.postRepo.GetPostsPersonalFeed(username, &lastPost, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Fill GeneralFeedDTO with posts
	for _, post := range posts {
		authorDto := models.AuthorDTO{
			Username:          post.User.Username,
			Nickname:          post.User.Nickname,
			ProfilePictureUrl: post.User.ProfilePictureUrl,
		}
		postDto := models.PostCreateResponseDTO{
			PostId:       post.Id,
			Author:       &authorDto,
			CreationDate: post.CreatedAt,
			Content:      post.Content,
		}
		feed.Records = append(feed.Records, postDto)
	}

	// Set pagination details
	if len(posts) > 0 {
		feed.Pagination.LastPostId = posts[len(posts)-1].Id
	}
	feed.Pagination.Limit = limit
	feed.Pagination.Records = totalPostsCount

	return &feed, nil, http.StatusOK
}
