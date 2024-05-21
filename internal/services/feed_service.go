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
	postRepo    repositories.PostRepositoryInterface
	userRepo    repositories.UserRepositoryInterface
	likeRepo    repositories.LikeRepositoryInterface
	commentRepo repositories.CommentRepositoryInterface
}

// NewFeedService can be used as a constructor to create a FeedService "object"
func NewFeedService(postRepo repositories.PostRepositoryInterface, userRepo repositories.UserRepositoryInterface, likeRepo repositories.LikeRepositoryInterface, commentRepo repositories.CommentRepositoryInterface) *FeedService {
	return &FeedService{postRepo: postRepo, userRepo: userRepo, likeRepo: likeRepo, commentRepo: commentRepo}
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
	postDtos := []models.UserFeedRecordDTO{} // no empty slice declaration using array literal (suggested by Goland) because it would be nil when marshalled to json instead of []
	for _, post := range posts {
		var locationDTO *models.LocationDTO

		likedByCurrentUser, likeCount, commentCount, err := service.getLikeAndCommentInformationByPost(post, currentUsername)
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

		repostDto, err := service.getRepostResponseDto(post, currentUsername)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}

		postDto := models.UserFeedRecordDTO{
			PostId:       post.Id.String(),
			CreationDate: post.CreatedAt,
			Content:      post.Content,
			Comments:     commentCount,
			Likes:        likeCount,
			Liked:        likedByCurrentUser,
			Location:     locationDTO,
			Repost:       repostDto,
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

		likedByCurrentUser, likeCount, commentCount, err := service.getLikeAndCommentInformationByPost(post, currentUsername)
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

		repostDto, err := service.getRepostResponseDto(post, currentUsername)
		if err != nil {
			return nil, err
		}

		postDto := models.PostResponseDTO{
			PostId:       post.Id,
			Author:       &authorDto,
			CreationDate: post.CreatedAt,
			Content:      post.Content,
			Comments:     commentCount,
			Likes:        likeCount,
			Liked:        likedByCurrentUser,
			Location:     locationDTO,
			Repost:       repostDto,
		}
		feed.Records = append(feed.Records, postDto)
	}
	return &feed, nil
}

// getLikeAndCommentInformationByPost returns whether the post is liked by the current user and the like count
func (service *FeedService) getLikeAndCommentInformationByPost(post models.Post, currentUsername string) (bool, int64, int64, error) {
	var likedByCurrentUser = false
	var likeCount int64
	var commentCount int64
	_, err := service.likeRepo.FindLike(post.Id.String(), currentUsername)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, 0, 0, err
	}
	if err == nil {
		likedByCurrentUser = true
	}
	likeCount, err = service.likeRepo.CountLikes(post.Id.String())
	if err != nil {
		return false, 0, 0, err
	}
	commentCount, err = service.commentRepo.CountComments(post.Id.String())
	if err != nil {
		return false, 0, 0, err
	}
	return likedByCurrentUser, likeCount, commentCount, nil
}

func (service *FeedService) getRepostResponseDto(post models.Post, currentUsername string) (*models.PostResponseDTO, error) {
	var repostDto *models.PostResponseDTO = nil

	// If post is not a repost, return empty dto
	if post.RepostId == nil {
		return repostDto, nil
	}

	// Get repost
	repost, err := service.postRepo.GetPostById(post.RepostId.String())
	if err != nil {

		// If repost is not found because it may have been deleted, return repost dto with only the repost id
		if errors.Is(err, gorm.ErrRecordNotFound) {
			repostDto = &models.PostResponseDTO{
				PostId: *post.RepostId,
			}
			return repostDto, nil
		}

		// Else database error
		return nil, err
	}

	// Get like information
	likedByCurrentUser, likeCount, commentCount, err := service.getLikeAndCommentInformationByPost(repost, currentUsername)
	if err != nil {
		return nil, err
	}

	// Create dto
	authorDto := models.AuthorDTO{
		Username:          repost.User.Username,
		Nickname:          repost.User.Nickname,
		ProfilePictureUrl: repost.User.ProfilePictureUrl,
	}

	var locationDTO *models.LocationDTO = nil
	if repost.LocationId != nil {
		locationDTO = &models.LocationDTO{
			Longitude: &repost.Location.Longitude,
			Latitude:  &repost.Location.Latitude,
			Accuracy:  &repost.Location.Accuracy,
		}
	}

	repostDto = &models.PostResponseDTO{
		PostId:       repost.Id,
		Author:       &authorDto,
		CreationDate: repost.CreatedAt,
		Content:      repost.Content,
		Comments:     commentCount,
		Likes:        likeCount,
		Liked:        likedByCurrentUser,
		Location:     locationDTO,
		Repost:       nil, // cannot have a repost of a repost, so always nil
	}
	return repostDto, nil
}
