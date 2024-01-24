package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type SubscriptionServiceInterface interface {
	PostSubscription(req *models.SubscriptionPostRequestDTO, currentUsername string) (*models.SubscriptionPostResponseDTO, *customerrors.CustomError, int)
	DeleteSubscription(subscriptionId string, currentUsername string) (*customerrors.CustomError, int)
	GetSubscriptions(ftype string, limit int, offset int, currentUsername string) (*models.SubscriptionSearchResponseDTO, *customerrors.CustomError, int)
}

type SubscriptionService struct {
	subscriptionRepo repositories.SubscriptionRepositoryInterface
	userRepo         repositories.UserRepositoryInterface
}

func NewSubscriptionService(
	subscriptionRepo repositories.SubscriptionRepositoryInterface,
	userRepo repositories.UserRepositoryInterface) *SubscriptionService {
	return &SubscriptionService{subscriptionRepo: subscriptionRepo, userRepo: userRepo}
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

func (service *SubscriptionService) GetSubscriptions(ftype string, limit int, offset int, username string) (*models.SubscriptionSearchResponseDTO, *customerrors.CustomError, int) {

	var followers []models.Subscription
	var followings []models.Subscription
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
	// prüfe ob Follower oder Followings abgefragt werden
	if ftype == "following" {
		//Ziehe Liste mit Benutzern, denen der User folgt
		followings, totalRecordsCount, err = service.subscriptionRepo.GetFollowings(limit, offset, username)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
		// Create response
		response := &models.SubscriptionSearchResponseDTO{
			Records: []models.SubscriptionSearchRecordDTO{},
			Pagination: &models.SubscriptionSearchPaginationDTO{
				Offset:  offset,
				Limit:   limit,
				Records: totalRecordsCount,
			},
		}

		for _, following := range followings {
			userDto := models.UserSearchRecordDTO{
				Username:          following.Following.Username,
				Nickname:          following.Following.Nickname,
				ProfilePictureUrl: following.Following.ProfilePictureUrl,
			}

			record := models.SubscriptionSearchRecordDTO{
				SubscriptionId:   following.Id,
				SubscriptionDate: following.SubscriptionDate,
				User:             userDto,
			}
			response.Records = append(response.Records, record)

		}
		return response, nil, http.StatusOK

	} else if ftype == "followers" {
		//Ziehe Liste mit Benutzern, die dem User folgen
		followers, totalRecordsCount, err = service.subscriptionRepo.GetFollowers(limit, offset, username)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
		// Create response
		response := &models.SubscriptionSearchResponseDTO{
			Records: []models.SubscriptionSearchRecordDTO{},
			Pagination: &models.SubscriptionSearchPaginationDTO{
				Offset:  offset,
				Limit:   limit,
				Records: totalRecordsCount,
			},
		}

		for _, follower := range followers {
			userDto := models.UserSearchRecordDTO{
				Username:          follower.Follower.Username,
				Nickname:          follower.Follower.Nickname,
				ProfilePictureUrl: follower.Follower.ProfilePictureUrl,
			}

			record := models.SubscriptionSearchRecordDTO{
				SubscriptionId:   follower.Id,
				SubscriptionDate: follower.SubscriptionDate,
				User:             userDto,
			}
			response.Records = append(response.Records, record)

		}
		return response, nil, http.StatusOK
	}
	return nil, customerrors.BadRequest, http.StatusBadRequest

}
