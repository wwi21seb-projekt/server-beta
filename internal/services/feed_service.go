package services

import (
	"errors"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
)

type FeedServiceInterface interface {
	GetPostsByUsername(username string, offset, limit int, currentUsername string) (*models.UserFeedDTO, *customerrors.CustomError, int)
	GetPostsGlobalFeed(lastPostId string, limit int, currentUsername string) (*models.GeneralFeedDTO, *customerrors.CustomError, int)
	GetPostsPersonalFeed(username string, lastPostId string, limit int, currentUsername string) (*models.GeneralFeedDTO, *customerrors.CustomError, int)
	GetPostsByHashtag(hashtag string, lastPostId string, limit int, currentUsername string) (*models.GeneralFeedDTO, *customerrors.CustomError, int)
}

type FeedService struct {
	postRepo repositories.PostRepositoryInterface
	userRepo repositories.UserRepositoryInterface
	likeRepo repositories.LikeRepositoryInterface
}

// NewFeedService can be used as a constructor to create a FeedService "object"
func NewFeedService(postRepo repositories.PostRepositoryInterface, userRepo repositories.UserRepositoryInterface, likeRepo repositories.LikeRepositoryInterface) *FeedService {
	return &FeedService{postRepo: postRepo, userRepo: userRepo, likeRepo: likeRepo}
}

// GetPostsByUsername returns a pagination object with the posts of a user using pagination parameters
func (service *FeedService) GetPostsByUsername(username string, offset, limit int, currentUsername string) (*models.UserFeedDTO, *customerrors.CustomError, int) {

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
	postDtos := []models.UserFeedRecordDTO{} // not using empty slice declaration (suggested by Goland) because it would be nil when marshalled to json instead of []
	for _, post := range posts {
		var locationDTO *models.LocationDTO

		likedByCurrentUser, likeCount, err := service.getLikeInformationByPost(post, currentUsername)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}

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
			Likes:        likeCount,
			Liked:        likedByCurrentUser,
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
func (service *FeedService) GetPostsGlobalFeed(lastPostId string, limit int, currentUsername string) (*models.GeneralFeedDTO, *customerrors.CustomError, int) {
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

		lastPost = post
	}

	// Retrieve posts from the database
	posts, totalPostsCount, err := service.postRepo.GetPostsGlobalFeed(&lastPost, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto
	feed, err := service.generatePostFeedWithAuthor(posts, totalPostsCount, limit, currentUsername)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	return feed, nil, http.StatusOK
}

// GetPostsPersonalFeed returns a pagination object with the posts in the personal feed using pagination parameters
func (service *FeedService) GetPostsPersonalFeed(username string, lastPostId string, limit int, currentUsername string) (*models.GeneralFeedDTO, *customerrors.CustomError, int) {
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

		lastPost = post
	}

	// Retrieve posts from the database
	posts, totalPostsCount, err := service.postRepo.GetPostsPersonalFeed(username, &lastPost, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto
	feed, err := service.generatePostFeedWithAuthor(posts, totalPostsCount, limit, currentUsername)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	return feed, nil, http.StatusOK
}

// GetPostsByHashtag returns a pagination object with the posts in the personal feed using pagination parameters
func (service *FeedService) GetPostsByHashtag(hashtag string, lastPostId string, limit int, currentUsername string) (*models.GeneralFeedDTO, *customerrors.CustomError, int) {
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

		lastPost = post
	}

	// Retrieve posts from the database
	posts, totalPostsCount, err := service.postRepo.GetPostsByHashtag(hashtag, &lastPost, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response dto
	feed, err := service.generatePostFeedWithAuthor(posts, totalPostsCount, limit, currentUsername)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	return feed, nil, http.StatusOK
}

// generatePostFeedWithAuthor creates a GeneralFeedDTO from a list of posts and a total count
func (service *FeedService) generatePostFeedWithAuthor(posts []models.Post, totalPostsCount int64, limit int, currentUsername string) (*models.GeneralFeedDTO, error) {
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

		likedByCurrentUser, likeCount, err := service.getLikeInformationByPost(post, currentUsername)
		if err != nil {
			return nil, err
		}

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
			Likes:        likeCount,
			Liked:        likedByCurrentUser,
			Location:     locationDTO,
		}
		feed.Records = append(feed.Records, postDto)
	}
	return &feed, nil
}

// getLikeInformationByPost returns whether the post is liked by the current user and the like count
func (service *FeedService) getLikeInformationByPost(post models.Post, currentUsername string) (bool, int64, error) {
	var likedByCurrentUser = false
	var likeCount int64
	_, err := service.likeRepo.FindLike(post.Id.String(), currentUsername)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, 0, err
	}
	if err == nil {
		likedByCurrentUser = true
	}
	likeCount = service.likeRepo.CountLikes(post.Id.String())
	return likedByCurrentUser, likeCount, nil
}
