package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"io"
	"net/http"
	"net/url"
	"os"
)

type PushSubscriptionServiceInterface interface {
	GetVapidKey() (*models.VapidKeyResponseDTO, *customerrors.CustomError, int)
	CreatePushSubscription(req *models.PushSubscriptionRequestDTO, currentUsername string) (*models.PushSubscriptionResponseDTO, *customerrors.CustomError, int)
	SendPushMessages(notificationObject interface{}, toUsername string)
}

type PushSubscriptionService struct {
	pushSubscriptionRepo repositories.PushSubscriptionRepositoryInterface

	vapidPrivateKey string
	vapidPublicKey  string
	serverMail      string
}

// NewPushSubscriptionService can be used as a constructor to create a PushSubscriptionService "object"
func NewPushSubscriptionService(pushSubscriptionRepo repositories.PushSubscriptionRepositoryInterface) *PushSubscriptionService {
	vapidPrivateKey := os.Getenv("VAPID_PRIVATE_KEY")
	vapidPublicKey := os.Getenv("VAPID_PUBLIC_KEY")
	serverMail := os.Getenv("EMAIL_ADDRESS")

	return &PushSubscriptionService{pushSubscriptionRepo: pushSubscriptionRepo, vapidPrivateKey: vapidPrivateKey, vapidPublicKey: vapidPublicKey, serverMail: serverMail}
}

// GetVapidKey returns a VAPID key for clients to register for push notifications
func (service *PushSubscriptionService) GetVapidKey() (*models.VapidKeyResponseDTO, *customerrors.CustomError, int) {
	// Return response object with public key
	response := models.VapidKeyResponseDTO{
		Key: service.vapidPublicKey,
	}
	return &response, nil, http.StatusOK
}

// CreatePushSubscription saves a new push subscription key to the database to send notifications to the client
func (service *PushSubscriptionService) CreatePushSubscription(req *models.PushSubscriptionRequestDTO, currentUsername string) (*models.PushSubscriptionResponseDTO, *customerrors.CustomError, int) {
	// Input validations
	// type either needs to be "web" or "expo"
	if req.Type != "web" && req.Type != "expo" {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	// endpoint needs to be a valid URL
	_, err := url.ParseRequestURI(req.SubscriptionInfo.Endpoint)
	if err != nil || len(req.SubscriptionInfo.Endpoint) <= 0 {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	// keys cannot be empty and need to be base64-URL encoded
	if len(req.SubscriptionInfo.SubscriptionKeys.P256dh) <= 0 || len(req.SubscriptionInfo.SubscriptionKeys.Auth) <= 0 {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	_, err = base64.RawURLEncoding.DecodeString(req.SubscriptionInfo.SubscriptionKeys.P256dh)
	if err != nil {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	_, err = base64.RawURLEncoding.DecodeString(req.SubscriptionInfo.SubscriptionKeys.Auth)
	if err != nil {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Create a new push subscription
	newPushSubscription := models.PushSubscription{
		Id:       uuid.New(),
		Username: currentUsername,
		Type:     req.Type,
		Endpoint: req.SubscriptionInfo.Endpoint,
		P256dh:   req.SubscriptionInfo.SubscriptionKeys.P256dh,
		Auth:     req.SubscriptionInfo.SubscriptionKeys.Auth,
	}

	err = service.pushSubscriptionRepo.CreatePushSubscription(&newPushSubscription)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response object with subscription id
	responseDto := models.PushSubscriptionResponseDTO{
		SubscriptionId: newPushSubscription.Id.String(),
	}

	return &responseDto, nil, http.StatusCreated
}

// SendPushMessages sends push messages to all push subscriptions of a user
func (service *PushSubscriptionService) SendPushMessages(notificationObject interface{}, toUsername string) {
	// Create notification json string from object
	notificationJson, err := json.Marshal(notificationObject)
	if err != nil {
		return
	}
	notificationString := string(notificationJson)

	// Get all push subscriptions by username
	pushSubscriptions, err := service.pushSubscriptionRepo.GetPushSubscriptionsByUsername(toUsername)
	if err != nil {
		return
	}

	// Send push messages
	for _, pushSubscription := range pushSubscriptions {
		go service.sendPushToClient(&pushSubscription, notificationString) // send push message in background
	}
}

func (service *PushSubscriptionService) sendPushToClient(pushSubscription *models.PushSubscription, notificationString string) {
	sub := &webpush.Subscription{
		Endpoint: pushSubscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: pushSubscription.P256dh,
			Auth:   pushSubscription.Auth,
		},
	}

	resp, err := webpush.SendNotification([]byte(notificationString), sub, &webpush.Options{
		Subscriber:      service.serverMail,
		VAPIDPublicKey:  service.vapidPublicKey,
		VAPIDPrivateKey: service.vapidPrivateKey,
		TTL:             30,
	})

	if err != nil {
		fmt.Println(err, ", error sending notification to", pushSubscription.Username)
		return
	}

	fmt.Println("Notification sent to", pushSubscription.Username)
	fmt.Println("Status:", resp.Status)
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println("Response Body:", string(bodyBytes))

	// If the subscription is deactivated or expired, delete it
	if resp.StatusCode == http.StatusGone {
		_ = service.pushSubscriptionRepo.DeletePushSubscriptionById(pushSubscription.Id.String())
		return
	}
}
