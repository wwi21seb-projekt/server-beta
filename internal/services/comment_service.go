package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

type CommentServiceInterface interface {
	CreateComment(req *models.CommentCreateRequestDTO, postId, currentUsername string) (*models.CommentCreateResponseDTO, *customerrors.CustomError, int)
	GetCommentsByPostId(postId string, offset, limit int) (*models.CommentFeedResponseDTO, *customerrors.CustomError, int)
}

type CommentService struct {
	commentRepo repositories.CommentRepositoryInterface
	postRepo    repositories.PostRepositoryInterface
	policy      *bluemonday.Policy
}

// NewCommentService can be used as a constructor to create a CommentService "object"
func NewCommentService(commentRepo repositories.CommentRepositoryInterface, postRepo repositories.PostRepositoryInterface) *CommentService {
	return &CommentService{commentRepo: commentRepo, postRepo: postRepo, policy: bluemonday.UGCPolicy()}
}

// CreateComment creates a new comment for a given post id using the provided request data
func (service *CommentService) CreateComment(req *models.CommentCreateRequestDTO, postId, currentUsername string) (*models.CommentCreateResponseDTO, *customerrors.CustomError, int) {
	// Sanitize content because it is a free text field
	req.Content = strings.Trim(req.Content, " ") // remove leading and trailing whitespaces
	req.Content = service.policy.Sanitize(req.Content)

	// Content must not be empty or exceed 128 characters
	if len(req.Content) <= 0 || len(req.Content) > 128 {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Post ID must be a valid UUID, otherwise post does not exist
	postIdUUID, err := uuid.Parse(postId)
	if err != nil {
		return nil, customerrors.PostNotFound, http.StatusNotFound
	}

	// Check if post exists
	_, err = service.postRepo.GetPostById(postId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.PostNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create comment
	comment := &models.Comment{
		Id:        uuid.New(),
		PostID:    postIdUUID,
		Username:  currentUsername,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	err = service.commentRepo.CreateComment(comment)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Prepare response
	responseDto := &models.CommentCreateResponseDTO{
		CommentId:    comment.Id,
		Content:      comment.Content,
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
	var commentRecords []models.CommentRecordDTO
	for _, comment := range comments {
		commentRecords = append(commentRecords, models.CommentRecordDTO{
			CommentId: comment.Id,
			Content:   comment.Content,
			Author: &models.AuthorDTO{
				Username:          comment.User.Username,
				Nickname:          comment.User.Nickname,
				ProfilePictureUrl: comment.User.ProfilePictureUrl,
			},
			CreationDate: comment.CreatedAt,
		})
	}

	paginationDto := &models.CommentPaginationDTO{
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
