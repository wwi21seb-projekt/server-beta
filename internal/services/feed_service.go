package services

import (
	"errors"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
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
	postDtos := make([]models.UserFeedRecordDTO, 0)
	for _, post := range posts {
		likedByCurrentUser, likeCount, commentCount, err := service.getLikeAndCommentInformationByPost(post, currentUsername)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
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
			Location:     utils.GenerateLocationDTOFromLocation(&post.Location),
			Picture:      utils.GenerateImageMetadataDTOFromImage(&post.Image),
			Repost:       repostDto,
		}
		postDtos = append(postDtos, postDto)
	}

	paginationDto := models.OffsetPaginationDTO{
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
				emptyFeed := service.createEmptyFeedObject(limit, totalPostsCount)
				return emptyFeed, nil, http.StatusOK
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
				emptyFeed := service.createEmptyFeedObject(limit, totalPostsCount)
				return emptyFeed, nil, http.StatusOK
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
				_, totalPostsCount, err := service.postRepo.GetPostsByHashtag(hashtag, &lastPost, limit)
				if err != nil {
					return nil, customerrors.DatabaseError, http.StatusInternalServerError
				}
				emptyFeed := service.createEmptyFeedObject(limit, totalPostsCount)
				return emptyFeed, nil, http.StatusOK
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
		Pagination: &models.PostCursorPaginationDTO{
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

		repostDto, err := service.getRepostResponseDto(post, currentUsername)
		if err != nil {
			return nil, err
		}

		postDto := models.PostResponseDTO{
			PostId:       post.Id,
			Author:       utils.GenerateUserDTOFromUser(&post.User),
			CreationDate: post.CreatedAt,
			Content:      post.Content,
			Picture:      utils.GenerateImageMetadataDTOFromImage(&post.Image),
			Comments:     commentCount,
			Likes:        likeCount,
			Liked:        likedByCurrentUser,
			Location:     utils.GenerateLocationDTOFromLocation(&post.Location),
			Repost:       repostDto,
		}
		feed.Records = append(feed.Records, postDto)
	}
	return &feed, nil
}

// getLikeAndCommentInformationByPost returns whether the post is liked by the current user and the like count
func (service *FeedService) getLikeAndCommentInformationByPost(post models.Post, currentUsername string) (bool, int64, int64, error) {
	var likedByCurrentUser = false
	_, err := service.likeRepo.FindLike(post.Id.String(), currentUsername)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, 0, 0, err
	}
	if err == nil {
		likedByCurrentUser = true
	}
	likeCount, err := service.likeRepo.CountLikes(post.Id.String())
	if err != nil {
		return false, 0, 0, err
	}
	commentCount, err := service.commentRepo.CountComments(post.Id.String())
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

		// If repost is not found because it may have been deleted, return empty repost dto with only the repost id
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

	repostDto = &models.PostResponseDTO{
		PostId:       repost.Id,
		Author:       utils.GenerateUserDTOFromUser(&repost.User),
		CreationDate: repost.CreatedAt,
		Content:      repost.Content,
		Comments:     commentCount,
		Picture:      utils.GenerateImageMetadataDTOFromImage(&repost.Image),
		Likes:        likeCount,
		Liked:        likedByCurrentUser,
		Location:     utils.GenerateLocationDTOFromLocation(&repost.Location),
		Repost:       nil, // cannot have a repost of a repost, so always nil
	}
	return repostDto, nil
}

// createEmptyFeedObject creates an empty feed object with the total number of records
func (service *FeedService) createEmptyFeedObject(limit int, totalPostsCount int64) *models.GeneralFeedDTO {
	emptyFeed := models.GeneralFeedDTO{
		Records: []models.PostResponseDTO{},
		Pagination: &models.PostCursorPaginationDTO{
			LastPostId: "",
			Limit:      limit,
			Records:    totalPostsCount,
		},
	}
	return &emptyFeed
}
