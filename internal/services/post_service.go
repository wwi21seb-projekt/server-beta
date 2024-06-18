package services

import (
	"encoding/base64"
	"errors"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

type PostServiceInterface interface {
	CreatePost(req *models.PostCreateRequestDTO, username string) (*models.PostResponseDTO, *customerrors.CustomError, int)
	DeletePost(postId string, username string) (*customerrors.CustomError, int)
}

type PostService struct {
	postRepo            repositories.PostRepositoryInterface
	userRepo            repositories.UserRepositoryInterface
	hashtagRepo         repositories.HashtagRepositoryInterface
	imageService        ImageServiceInterface
	validator           utils.ValidatorInterface
	likeRepo            repositories.LikeRepositoryInterface
	commentRepo         repositories.CommentRepositoryInterface
	policy              *bluemonday.Policy
	notificationService NotificationServiceInterface
}

// NewPostService can be used as a constructor to create a PostService "object"
func NewPostService(postRepo repositories.PostRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	hashtagRepo repositories.HashtagRepositoryInterface,
	validator utils.ValidatorInterface,
	likeRepo repositories.LikeRepositoryInterface,
	commentRepo repositories.CommentRepositoryInterface,
	notificationService NotificationServiceInterface) *PostService {
	return &PostService{postRepo: postRepo, userRepo: userRepo, hashtagRepo: hashtagRepo, validator: validator, likeRepo: likeRepo, commentRepo: commentRepo, policy: bluemonday.UGCPolicy(), notificationService: notificationService}
}

func (service *PostService) CreatePost(req *models.PostCreateRequestDTO, username string) (*models.PostResponseDTO, *customerrors.CustomError, int) {
	// Sanitize content because it is a free text field
	// Other fields are checked with regex patterns, that don't allow for malicious input
	req.Content = strings.Trim(req.Content, " ") // remove leading and trailing whitespaces
	req.Content = service.policy.Sanitize(req.Content)

	// Get repost if a repost id is given
	var repostDto *models.PostResponseDTO
	var repostId *uuid.UUID
	if req.RepostedPostId != "" {
		repost, err := service.postRepo.GetPostById(req.RepostedPostId)
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
		repostDto = createPostResponseFromPostObject(&repost, &repost.User, &repost.Location, &repost.Image, repostDto, repostCommentsCount, repostLikeCount, repostLikedByCurrentUser)
	}

	// Validations: 0-256 characters and utf8 characters
	if len(req.Content) > 256 {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	if len(req.Content) <= 0 && req.Picture == "" && req.RepostedPostId == "" { // either content, repostId or image is required
		return nil, customerrors.BadRequest, http.StatusBadRequest // location is neither necessary nor sufficient
	}
	if !utf8.ValidString(req.Content) {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Validate and create image object
	var image *models.Image
	var imageId *uuid.UUID
	if req.Picture != "" {
		imageBytes, err := base64.StdEncoding.DecodeString(req.Picture)
		if err != nil {
			return nil, customerrors.BadRequest, http.StatusBadRequest
		}
		valid, format, width, height := service.validator.ValidateImage(imageBytes)
		if !valid {
			return nil, customerrors.BadRequest, http.StatusBadRequest
		}
		image = &models.Image{
			Id:        uuid.New(),
			Format:    format,
			ImageData: imageBytes,
			Width:     width,
			Height:    height,
			Tag:       time.Now(),
		}
		imageId = &image.Id
	}

	// Create location
	var location *models.Location
	if req.Location != nil { // if location is present, create location object and save it
		// Check if coordinates are in valid range
		// Accuracy is bound to unsigned int, so no extra validation needed
		if !service.validator.ValidateLongitude(*req.Location.Longitude) || !service.validator.ValidateLatitude(*req.Location.Latitude) {
			return nil, customerrors.BadRequest, http.StatusBadRequest
		}

		location = &models.Location{
			Id:        uuid.New(),
			Longitude: *req.Location.Longitude,
			Latitude:  *req.Location.Latitude,
			Accuracy:  *req.Location.Accuracy,
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

	// Create post
	post := models.Post{
		Id:        uuid.New(),
		Username:  username,
		Content:   req.Content,
		Hashtags:  hashtags,
		CreatedAt: time.Now(),
		RepostId:  repostId,
	}

	// Add image to post if image was given
	if imageId != nil {
		post.ImageId = imageId
		post.Image = *image
	}
	// Add location to post if location was given
	if location != nil {
		post.LocationId = &location.Id
		post.Location = *location
	}

	// Save post to database
	err = service.postRepo.CreatePost(&post)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto and return
	postDto := createPostResponseFromPostObject(&post, user, location, image, repostDto, 0, 0, false) // no likes and comments yet

	// Create notification for owner of original post
	if repostId != nil {
		_ = service.notificationService.CreateNotification("repost", repostDto.Author.Username, username)
	}
	return postDto, nil, http.StatusCreated
}

func createPostResponseFromPostObject(
	post *models.Post, user *models.User,
	location *models.Location,
	image *models.Image,
	repostDto *models.PostResponseDTO,
	commentsCount int64,
	likesCount int64,
	likedByCurrentUser bool) *models.PostResponseDTO {
	// Create sub dto objects
	var locationDto *models.LocationDTO
	if post.LocationId != nil {
		locationDto = &models.LocationDTO{
			Longitude: &location.Longitude,
			Latitude:  &location.Latitude,
			Accuracy:  &location.Accuracy,
		}
	}
	var imageDto *models.ImageMetadataDTO
	if post.ImageId != nil {
		imageDto = &models.ImageMetadataDTO{
			Url:    utils.FormatImageUrl(post.ImageId.String(), image.Format),
			Width:  image.Width,
			Height: image.Height,
			Tag:    image.Tag,
		}
	}
	var userImageDto *models.ImageMetadataDTO
	if user.ImageId != nil {
		userImageDto = &models.ImageMetadataDTO{
			Url:    utils.FormatImageUrl(user.ImageId.String(), user.Image.Format),
			Width:  user.Image.Width,
			Height: user.Image.Height,
			Tag:    user.Image.Tag,
		}
	}

	// Create post dto
	postDto := models.PostResponseDTO{
		PostId: post.Id,
		Author: &models.UserDTO{
			Username: user.Username,
			Nickname: user.Nickname,
			Picture:  userImageDto,
		},
		CreationDate: post.CreatedAt,
		Content:      post.Content,
		Picture:      imageDto,
		Comments:     commentsCount,
		Likes:        likesCount,
		Liked:        likedByCurrentUser,
		Location:     locationDto,
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
