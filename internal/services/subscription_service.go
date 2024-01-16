package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type SubscriptionServiceInterface interface {
	PostSubscription(req *models.SubscriptionPostRequestDTO, currentUsername string) (*models.SubscriptionPostResponseDTO, *customerrors.CustomError, int)
	DeleteSubscription(subscriptionId string, currentUsername string) (*customerrors.CustomError, int)
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
		return nil, customerrors.PreliminarySelfFollow, http.StatusNotAcceptable
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
		return nil, customerrors.PreliminarySubscriptionAlreadyExists, http.StatusConflict
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
			return customerrors.PreliminarySubscriptionNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if user is authorized to delete subscription
	if subscription.FollowerUsername != currentUsername {
		return customerrors.PreliminarySubscriptionDeleteNotAuthorized, http.StatusForbidden
	}

	// Delete subscription
	err = service.subscriptionRepo.DeleteSubscription(subscriptionId)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}
