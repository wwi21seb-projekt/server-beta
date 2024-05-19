package services

import (
	"errors"
	"fmt"
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
	GetPostsByUsername(username string, offset, limit int) (*models.UserFeedDTO, *customerrors.CustomError, int)
	GetPostsGlobalFeed(lastPostId string, limit int) (*models.GeneralFeedDTO, *customerrors.CustomError, int)
	GetPostsPersonalFeed(username string, lastPostId string, limit int) (*models.GeneralFeedDTO, *customerrors.CustomError, int)
	DeletePost(postId string, username string) (*customerrors.CustomError, int)
	GetPostsByHashtag(hashtag string, lastPostId string, limit int) (*models.GeneralFeedDTO, *customerrors.CustomError, int)
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
	if len(req.Content) <= 0 && file == nil && len(req.RepostedPostId) <= 0 { // image or file can be empty, or it can be a repost, but not none
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
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Check if user is activated
	if !user.Activated {
		return nil, customerrors.UserUnauthorized, http.StatusUnauthorized
	}

	// check repost and if repost is already a repost

	var repost *models.Post = nil
	if len(req.RepostedPostId) > 0 {
		repostObj, err := service.postRepo.GetPostById(req.RepostedPostId)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
		if repostObj.RepostedPostId != nil {
			return nil, customerrors.BadRequest, http.StatusBadRequest
		}
		repost = repostObj // Ensure we correctly convert the value to a pointer
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
		Id:             uuid.New(),
		Username:       username,
		Content:        req.Content,
		ImageUrl:       imageUrl,
		Hashtags:       hashtags,
		CreatedAt:      time.Now(),
		RepostedPostId: nil,
		LocationId:     locationId,
	}
	if repost != nil {
		post.RepostedPostId = &repost.Id
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

	// create repostDTO if it's a repost
	var repostDTO *models.RepostDTO
	if repost != nil {
		repostDTO = &models.RepostDTO{

			Author: &models.AuthorDTO{
				Username:          repost.User.Username,
				Nickname:          repost.User.Nickname,
				ProfilePictureUrl: repost.User.ProfilePictureUrl,
			},
			CreationDate: repost.CreatedAt,
			Content:      repost.Content,
			Location: &models.LocationDTO{
				Longitude: &repost.Location.Longitude,
				Latitude:  &repost.Location.Latitude,
				Accuracy:  &repost.Location.Accuracy},
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
		Repost:       repostDTO,
		Location:     locationDTO,
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
		var locationDTO *models.LocationDTO
		if post.LocationId != nil {
			tempLatitude := post.Location.Latitude // need to use temp variables because the pointers change in the loop
			tempLongitude := post.Location.Longitude
			tempAccuracy := post.Location.Accuracy
			locationDTO = &models.LocationDTO{
				Longitude: &tempLongitude,
				Latitude:  &tempLatitude,
				Accuracy:  &tempAccuracy,
			}
		}
		postDto := models.UserFeedRecordDTO{
			PostId:       post.Id.String(),
			CreationDate: post.CreatedAt,
			Content:      post.Content,
			Location:     locationDTO,
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
	// Get last post if lastPostId is not empty
	var lastPost models.Post
	if lastPostId != "" {
		post, err := service.postRepo.GetPostById(lastPostId)
		if err != nil {

			// If post is not found, return empty feed with number of records
			if errors.Is(err, gorm.ErrRecordNotFound) {
				_, totalPostsCount, err := service.postRepo.GetPostsGlobalFeed(&lastPost, limit)
				if err != nil {
					return nil, customerrors.DatabaseError, http.StatusInternalServerError
				}
				emptyFeed := models.GeneralFeedDTO{
					Records: []models.PostResponseDTO{},
					Pagination: &models.GeneralFeedPaginationDTO{
						LastPostId: "",
						Limit:      limit,
						Records:    totalPostsCount,
					},
				}
				return &emptyFeed, nil, http.StatusOK
			}

			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}

		lastPost = *post
	}

	// Retrieve posts from the database
	posts, totalPostsCount, err := service.postRepo.GetPostsGlobalFeed(&lastPost, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto
	feed := generatePostFeedWithAuthor(posts, totalPostsCount, limit)

	return feed, nil, http.StatusOK
}

// GetPostsPersonalFeed returns a pagination object with the posts in the personal feed using pagination parameters
func (service *PostService) GetPostsPersonalFeed(username string, lastPostId string, limit int) (*models.GeneralFeedDTO, *customerrors.CustomError, int) {
	// Get last post if lastPostId is not empty
	var lastPost models.Post
	if lastPostId != "" {
		post, err := service.postRepo.GetPostById(lastPostId)
		if err != nil {

			// If post is not found, return empty feed with number of records
			if errors.Is(err, gorm.ErrRecordNotFound) {
				_, totalPostsCount, err := service.postRepo.GetPostsPersonalFeed(username, &lastPost, limit)
				if err != nil {
					return nil, customerrors.DatabaseError, http.StatusInternalServerError
				}
				emptyFeed := models.GeneralFeedDTO{
					Records: []models.PostResponseDTO{},
					Pagination: &models.GeneralFeedPaginationDTO{
						LastPostId: "",
						Limit:      limit,
						Records:    totalPostsCount,
					},
				}
				return &emptyFeed, nil, http.StatusOK
			}

			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}

		lastPost = *post
	}

	// Retrieve posts from the database
	posts, totalPostsCount, err := service.postRepo.GetPostsPersonalFeed(username, &lastPost, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto
	feed := generatePostFeedWithAuthor(posts, totalPostsCount, limit)

	return feed, nil, http.StatusOK
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

// GetPostsByHashtag returns a pagination object with the posts in the personal feed using pagination parameters
func (service *PostService) GetPostsByHashtag(hashtag string, lastPostId string, limit int) (*models.GeneralFeedDTO, *customerrors.CustomError, int) {
	// Get last post if lastPostId is not empty
	var lastPost models.Post
	if lastPostId != "" {
		post, err := service.postRepo.GetPostById(lastPostId)
		if err != nil {

			// If post is not found, return empty feed with number of records
			if errors.Is(err, gorm.ErrRecordNotFound) {
				_, totalPostsCount, err := service.postRepo.GetPostsGlobalFeed(&lastPost, limit)
				if err != nil {
					return nil, customerrors.DatabaseError, http.StatusInternalServerError
				}
				emptyFeed := models.GeneralFeedDTO{
					Records: []models.PostResponseDTO{},
					Pagination: &models.GeneralFeedPaginationDTO{
						LastPostId: "",
						Limit:      limit,
						Records:    totalPostsCount,
					},
				}
				return &emptyFeed, nil, http.StatusOK
			}

			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}

		lastPost = *post
	}

	// Retrieve posts from the database
	posts, totalPostsCount, err := service.postRepo.GetPostsByHashtag(hashtag, &lastPost, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto
	feed := generatePostFeedWithAuthor(posts, totalPostsCount, limit)

	return feed, nil, http.StatusOK
}

// generatePostFeedWithAuthor creates a GeneralFeedDTO from a list of posts and a total count
func generatePostFeedWithAuthor(posts []models.Post, totalPostsCount int64, limit int) *models.GeneralFeedDTO {
	// Create response dto
	newLastPostId := ""
	if len(posts) > 0 {
		newLastPostId = posts[len(posts)-1].Id.String()
	}
	feed := models.GeneralFeedDTO{
		Records: []models.PostResponseDTO{},
		Pagination: &models.GeneralFeedPaginationDTO{
			LastPostId: newLastPostId,
			Limit:      limit,
			Records:    totalPostsCount,
		},
	}
	for _, post := range posts {
		authorDto := models.AuthorDTO{
			Username:          post.User.Username,
			Nickname:          post.User.Nickname,
			ProfilePictureUrl: post.User.ProfilePictureUrl,
		}
		var locationDTO *models.LocationDTO = nil
		if post.LocationId != nil {
			tempLatitude := post.Location.Latitude // need to use temp variables because the pointers change in the loop
			tempLongitude := post.Location.Longitude
			tempAccuracy := post.Location.Accuracy
			locationDTO = &models.LocationDTO{
				Longitude: &tempLongitude,
				Latitude:  &tempLatitude,
				Accuracy:  &tempAccuracy,
			}
		}
		postDto := models.PostResponseDTO{
			PostId:       post.Id,
			Author:       &authorDto,
			CreationDate: post.CreatedAt,
			Content:      post.Content,
			Location:     locationDTO,
		}
		if locationDTO != nil {
			fmt.Println("Post id: ", post.Id)
			fmt.Println("Location: ", post.Location.Longitude, post.Location.Latitude, post.Location.Accuracy)
			fmt.Println("DTO: ", *locationDTO.Longitude, *locationDTO.Latitude, *locationDTO.Accuracy)
		}
		feed.Records = append(feed.Records, postDto)
	}
	return &feed
}
