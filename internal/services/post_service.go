package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

type PostServiceInterface interface {
	CreatePost(req *models.PostCreateRequestDTO, file *multipart.FileHeader, username string) (*models.PostResponseDTO, *customerrors.CustomError, int)
	DeletePost(postId string, username string) (*customerrors.CustomError, int)
}

type PostService struct {
	postRepo            repositories.PostRepositoryInterface
	userRepo            repositories.UserRepositoryInterface
	hashtagRepo         repositories.HashtagRepositoryInterface
	imageService        ImageServiceInterface
	validator           utils.ValidatorInterface
	locationRepo        repositories.LocationRepositoryInterface
	likeRepo            repositories.LikeRepositoryInterface
	commentRepo         repositories.CommentRepositoryInterface
	policy              *bluemonday.Policy
	notificationService NotificationServiceInterface
}

// NewPostService can be used as a constructor to create a PostService "object"
func NewPostService(postRepo repositories.PostRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	hashtagRepo repositories.HashtagRepositoryInterface,
	imageService ImageServiceInterface,
	validator utils.ValidatorInterface,
	locationRepo repositories.LocationRepositoryInterface,
	likeRepo repositories.LikeRepositoryInterface,
	commentRepo repositories.CommentRepositoryInterface,
	notificationService NotificationServiceInterface) *PostService {
	return &PostService{postRepo: postRepo, userRepo: userRepo, hashtagRepo: hashtagRepo, imageService: imageService, validator: validator, locationRepo: locationRepo, likeRepo: likeRepo, commentRepo: commentRepo, policy: bluemonday.UGCPolicy(), notificationService: notificationService}
}

func (service *PostService) CreatePost(req *models.PostCreateRequestDTO, file *multipart.FileHeader, username string) (*models.PostResponseDTO, *customerrors.CustomError, int) {
	// Sanitize content because it is a free text field
	// Other fields are checked with regex patterns, that don't allow for malicious input
	req.Content = strings.Trim(req.Content, " ") // remove leading and trailing whitespaces
	req.Content = service.policy.Sanitize(req.Content)

	// Get repost if a repost id is given
	var repostDto *models.PostResponseDTO
	var repostId *uuid.UUID
	if req.RepostId != nil {
		repost, err := service.postRepo.GetPostById(*req.RepostId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, customerrors.PostNotFound, http.StatusNotFound
			}
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
		repostId = &repost.Id

		// Check if repost is a repost
		if repost.RepostId != nil {
			return nil, customerrors.BadRequest, http.StatusBadRequest // repost of a repost is not allowed
		}

		// Get like and comments information of repost
		var repostLikeCount int64 = 0
		var repostLikedByCurrentUser = false
		var repostCommentsCount int64 = 0
		_, err = service.likeRepo.FindLike(repost.Id.String(), username)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
		if err == nil {
			repostLikedByCurrentUser = true
		}
		repostLikeCount, err = service.likeRepo.CountLikes(repost.Id.String())
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
		repostCommentsCount, err = service.commentRepo.CountComments(repost.Id.String())
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}

		// Create dto
		repostDto = createPostResponseFromPostObject(&repost, &repost.User, &repost.Location, nil, repostCommentsCount, repostLikeCount, repostLikedByCurrentUser)

	}

	// Validations: 0-256 characters and utf8 characters
	if len(req.Content) > 256 {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	if len(req.Content) <= 0 && file == nil && req.RepostId == nil { // either content, repostId or image is required
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
			return nil, customerrors.UserUnauthorized, http.StatusUnauthorized // not reachable, because of JWT middleware
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	//Extract hashtags
	hashtagNames := utils.ExtractHashtags(req.Content)

	// Create hashtag object for each hashtag name
	var hashtags []models.Hashtag
	for _, name := range hashtagNames {
		hashtag, err := service.hashtagRepo.FindOrCreateHashtag(name)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
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
	var location *models.Location
	if req.Location != nil { // if location is present, create location object and save it
		location = &models.Location{
			Id:        uuid.New(),
			Longitude: *req.Location.Longitude,
			Latitude:  *req.Location.Latitude,
			Accuracy:  *req.Location.Accuracy,
		}
		locationId = &location.Id
		err = service.locationRepo.CreateLocation(location)
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
		RepostId:   repostId,
	}
	err = service.postRepo.CreatePost(&post)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto and return
	postDto := createPostResponseFromPostObject(&post, user, location, repostDto, 0, 0, false) // no likes and comments yet

	// Create notification for owner of original post
	if repostId != nil {
		_ = service.notificationService.CreateNotification("repost", repostDto.Author.Username, username)
	}
	return postDto, nil, http.StatusCreated
}

func createPostResponseFromPostObject(
	post *models.Post, user *models.User,
	location *models.Location,
	repostDto *models.PostResponseDTO,
	commentsCount int64,
	likesCount int64,
	likedByCurrentUser bool) *models.PostResponseDTO {
	var locationDTO *models.LocationDTO
	if post.LocationId != nil {
		locationDTO = &models.LocationDTO{
			Longitude: &location.Longitude,
			Latitude:  &location.Latitude,
			Accuracy:  &location.Accuracy,
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
		Comments:     commentsCount,
		Likes:        likesCount,
		Liked:        likedByCurrentUser,
		Location:     locationDTO,
		Repost:       repostDto,
	}
	return &postDto
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
