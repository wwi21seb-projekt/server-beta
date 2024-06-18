package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type SubscriptionServiceInterface interface {
	PostSubscription(req *models.SubscriptionPostRequestDTO, currentUsername string) (*models.SubscriptionPostResponseDTO, *customerrors.CustomError, int)
	DeleteSubscription(subscriptionId string, currentUsername string) (*customerrors.CustomError, int)
	GetSubscriptions(queryType string, limit int, offset int, username string, currentUsername string) (*models.SubscriptionResponseDTO, *customerrors.CustomError, int)
}

type SubscriptionService struct {
	subscriptionRepo    repositories.SubscriptionRepositoryInterface
	userRepo            repositories.UserRepositoryInterface
	notificationService NotificationServiceInterface
}

func NewSubscriptionService(
	subscriptionRepo repositories.SubscriptionRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	notificationService NotificationServiceInterface) *SubscriptionService {
	return &SubscriptionService{subscriptionRepo: subscriptionRepo, userRepo: userRepo, notificationService: notificationService}
}

func (service *SubscriptionService) PostSubscription(req *models.SubscriptionPostRequestDTO, currentUsername string) (*models.SubscriptionPostResponseDTO, *customerrors.CustomError, int) {

	// Check if user wants to follow himself
	if req.Following == currentUsername {
		return nil, customerrors.SelfFollow, http.StatusNotAcceptable
	}

	// Check if user exists
	_, err := service.userRepo.FindUserByUsername(req.Following)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if subscription already exists
	_, err = service.subscriptionRepo.GetSubscriptionByUsernames(currentUsername, req.Following)
	if err == nil {
		return nil, customerrors.SubscriptionAlreadyExists, http.StatusConflict
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create subscription
	newSubscription := models.Subscription{
		Id:                uuid.New(),
		SubscriptionDate:  time.Now(),
		FollowerUsername:  currentUsername,
		FollowingUsername: req.Following,
	}

	err = service.subscriptionRepo.CreateSubscription(&newSubscription)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create notification
	_ = service.notificationService.CreateNotification("follow", req.Following, currentUsername)

	// Create response
	response := &models.SubscriptionPostResponseDTO{
		SubscriptionId:   newSubscription.Id,
		SubscriptionDate: newSubscription.SubscriptionDate,
		Follower:         newSubscription.FollowerUsername,
		Following:        newSubscription.FollowingUsername,
	}
	return response, nil, http.StatusCreated
}

func (service *SubscriptionService) DeleteSubscription(subscriptionId string, currentUsername string) (*customerrors.CustomError, int) {

	// Get subscription
	subscription, err := service.subscriptionRepo.GetSubscriptionById(subscriptionId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.SubscriptionNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if user is authorized to delete subscription
	if subscription.FollowerUsername != currentUsername {
		return customerrors.SubscriptionDeleteNotAuthorized, http.StatusForbidden
	}

	// Delete subscription
	err = service.subscriptionRepo.DeleteSubscription(subscriptionId)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}

func (service *SubscriptionService) GetSubscriptions(queryType string, limit int, offset int, username string, currentUsername string) (*models.SubscriptionResponseDTO, *customerrors.CustomError, int) {

	var sqlRecords []models.UserSubscriptionSQLRecordDTO
	var totalRecordsCount int64
	var err error
	var _ *models.User

	_, err = service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if following or followers was requested
	if queryType == "following" {
		// Get list of users that the user follows
		sqlRecords, totalRecordsCount, err = service.subscriptionRepo.GetFollowings(limit, offset, username, currentUsername)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}

	} else if queryType == "followers" {
		// Get list of users that follow the user
		sqlRecords, totalRecordsCount, err = service.subscriptionRepo.GetFollowers(limit, offset, username, currentUsername)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
	}

	// Create response
	records := make([]models.UserSubscriptionRecordDTO, 0)

	for _, sqlRecord := range sqlRecords {
		var imageDto *models.ImageMetadataDTO
		if sqlRecord.ImageId != "" { // create metadata dto only if image exists
			imageDto = &models.ImageMetadataDTO{
				Url:    utils.FormatImageUrl(sqlRecord.ImageId, sqlRecord.Format),
				Width:  sqlRecord.Width,
				Height: sqlRecord.Height,
				Tag:    sqlRecord.Tag,
			}
		}
		record := models.UserSubscriptionRecordDTO{
			FollowerId:  sqlRecord.FollowerId,
			FollowingId: sqlRecord.FollowingId,
			Username:    sqlRecord.Username,
			Nickname:    sqlRecord.Nickname,
			Picture:     imageDto,
		}
		records = append(records, record)
	}

	response := &models.SubscriptionResponseDTO{
		Records: records,
		Pagination: &models.OffsetPaginationDTO{
			Offset:  offset,
			Limit:   limit,
			Records: totalRecordsCount,
		},
	}

	return response, nil, http.StatusOK
}
