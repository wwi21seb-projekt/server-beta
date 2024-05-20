package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"html"
	"mime/multipart"
	"net/http"
	"time"
	"unicode/utf8"
)

type PostServiceInterface interface {
	CreatePost(req *models.PostCreateRequestDTO, file *multipart.FileHeader, username string) (*models.PostResponseDTO, *customerrors.CustomError, int)
	DeletePost(postId string, username string) (*customerrors.CustomError, int)
}

type PostService struct {
	postRepo     repositories.PostRepositoryInterface
	userRepo     repositories.UserRepositoryInterface
	hashtagRepo  repositories.HashtagRepositoryInterface
	imageService ImageServiceInterface
	validator    utils.ValidatorInterface
	locationRepo repositories.LocationRepositoryInterface
}

// NewPostService can be used as a constructor to create a PostService "object"
func NewPostService(postRepo repositories.PostRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	hashtagRepo repositories.HashtagRepositoryInterface,
	imageService ImageServiceInterface,
	validator utils.ValidatorInterface,
	locationRepo repositories.LocationRepositoryInterface) *PostService {
	return &PostService{postRepo: postRepo, userRepo: userRepo, hashtagRepo: hashtagRepo, imageService: imageService, validator: validator, locationRepo: locationRepo}
}

func (service *PostService) CreatePost(req *models.PostCreateRequestDTO, file *multipart.FileHeader, username string) (*models.PostResponseDTO, *customerrors.CustomError, int) {
	// Escape html content to prevent XSS
	req.Content = html.EscapeString(req.Content)

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
	if req.Location != nil {
		// Then check if coordinates are in valid range
		if !service.validator.ValidateLongitude(*req.Location.Longitude) || !service.validator.ValidateLatitude(*req.Location.Latitude) {
			return nil, customerrors.BadRequest, http.StatusBadRequest
		}
	}

	// Get user
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserUnauthorized, http.StatusUnauthorized
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if user is activated
	if !user.Activated {
		return nil, customerrors.UserUnauthorized, http.StatusUnauthorized
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
	var locationId *uuid.UUID
	if req.Location != nil { // if location is present, create location object and save it
		location := models.Location{
			Id:        uuid.New(),
			Longitude: *req.Location.Longitude,
			Latitude:  *req.Location.Latitude,
			Accuracy:  *req.Location.Accuracy,
		}
		locationId = &location.Id
		err = service.locationRepo.CreateLocation(&location)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
	}

	post := models.Post{
		Id:         uuid.New(),
		Username:   username,
		Content:    req.Content,
		ImageUrl:   imageUrl,
		Hashtags:   hashtags,
		CreatedAt:  time.Now(),
		LocationId: locationId,
	}
	err = service.postRepo.CreatePost(&post)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto and return
	var locationDTO *models.LocationDTO
	if req.Location != nil {
		locationDTO = &models.LocationDTO{
			Longitude: req.Location.Longitude,
			Latitude:  req.Location.Latitude,
			Accuracy:  req.Location.Accuracy,
		}
	}
	postDto := models.PostResponseDTO{
		PostId: post.Id,
		Author: &models.AuthorDTO{
			Username:          user.Username,
			Nickname:          user.Nickname,
			ProfilePictureUrl: user.ProfilePictureUrl,
		},
		CreationDate: post.CreatedAt,
		Content:      post.Content,
		Likes:        0, // no likes yet
		Liked:        false,
		Location:     locationDTO,
	}

	return &postDto, nil, http.StatusCreated
}

// DeletePost deletes a post by id and returns an error if the post does not exist or the requesting user is not the author
func (service *PostService) DeletePost(postId string, username string) (*customerrors.CustomError, int) {
	// Find post by ID
	post, err := service.postRepo.GetPostById(postId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.PostNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if the requesting user is the author of the post
	if post.Username != username {
		return customerrors.PostDeleteForbidden, http.StatusForbidden
	}

	// Delete post
	err = service.postRepo.DeletePostById(postId)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}
