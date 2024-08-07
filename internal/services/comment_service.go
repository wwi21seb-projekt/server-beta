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
	"net/http"
	"strings"
	"time"
)

type CommentServiceInterface interface {
	CreateComment(req *models.CommentCreateRequestDTO, postId, currentUsername string) (*models.CommentResponseDTO, *customerrors.CustomError, int)
	GetCommentsByPostId(postId string, offset, limit int) (*models.CommentFeedResponseDTO, *customerrors.CustomError, int)
}

type CommentService struct {
	commentRepo repositories.CommentRepositoryInterface
	postRepo    repositories.PostRepositoryInterface
	userRepo    repositories.UserRepositoryInterface
	policy      *bluemonday.Policy
}

// NewCommentService can be used as a constructor to create a CommentService "object"
func NewCommentService(commentRepo repositories.CommentRepositoryInterface, postRepo repositories.PostRepositoryInterface, userRepo repositories.UserRepositoryInterface) *CommentService {
	return &CommentService{commentRepo: commentRepo, postRepo: postRepo, userRepo: userRepo, policy: bluemonday.UGCPolicy()}
}

// CreateComment creates a new comment for a given post id using the provided request data
func (service *CommentService) CreateComment(req *models.CommentCreateRequestDTO, postId, currentUsername string) (*models.CommentResponseDTO, *customerrors.CustomError, int) {
	// Sanitize content because it is a free text field
	req.Content = strings.Trim(req.Content, " ") // remove leading and trailing whitespaces
	req.Content = service.policy.Sanitize(req.Content)

	// Content must not be empty or exceed 128 characters
	if len(req.Content) <= 0 || len(req.Content) > 128 {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Check if post exists
	post, err := service.postRepo.GetPostById(postId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.PostNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Get user by username
	user, err := service.userRepo.FindUserByUsername(currentUsername)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.Unauthorized, http.StatusUnauthorized // not reachable, because of JWT middleware
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create comment
	comment := &models.Comment{
		Id:        uuid.New(),
		PostID:    post.Id,
		Username:  currentUsername,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	err = service.commentRepo.CreateComment(comment)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Prepare response
	responseDto := &models.CommentResponseDTO{
		CommentId:    comment.Id,
		Content:      comment.Content,
		Author:       utils.GenerateUserDTOFromUser(user),
		CreationDate: comment.CreatedAt,
	}

	return responseDto, nil, http.StatusCreated

}

// GetCommentsByPostId retrieves comments for a given post id using the provided pagination information
func (service *CommentService) GetCommentsByPostId(postId string, offset, limit int) (*models.CommentFeedResponseDTO, *customerrors.CustomError, int) {
	// Check if post exists
	_, err := service.postRepo.GetPostById(postId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.PostNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Get comments using pagination information
	comments, count, err := service.commentRepo.GetCommentsByPostId(postId, offset, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Prepare response
	var commentRecords []models.CommentResponseDTO
	for _, comment := range comments {
		commentRecords = append(commentRecords, models.CommentResponseDTO{
			CommentId:    comment.Id,
			Content:      comment.Content,
			Author:       utils.GenerateUserDTOFromUser(&comment.User),
			CreationDate: comment.CreatedAt,
		})
	}

	paginationDto := &models.OffsetPaginationDTO{
		Offset:  offset,
		Limit:   limit,
		Records: count,
	}

	responseDto := &models.CommentFeedResponseDTO{
		Records:    commentRecords,
		Pagination: paginationDto,
	}

	return responseDto, nil, http.StatusOK
}
