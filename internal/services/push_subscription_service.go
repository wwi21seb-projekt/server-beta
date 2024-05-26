package services

import (
	"bytes"
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
	"regexp"
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

	// For web push notifications endpoint, p256dh and auth keys are required
	if req.Type == "web" {
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

		req.Token = "" // token is not required for web push notifications
	}

	// For expo push notifications only token is required
	if req.Type == "expo" {
		// token needs to be in the format ExponentPushToken[...]
		tokenRegex := `ExponentPushToken\[[a-zA-Z0-9-_]+\]`
		if match, _ := regexp.MatchString(tokenRegex, req.Token); !match {
			return nil, customerrors.BadRequest, http.StatusBadRequest
		}

		req.SubscriptionInfo = &models.SubscriptionInfo{} // subscription info is not required for expo push notifications
	}

	// Create a new push subscription
	newPushSubscription := models.PushSubscription{
		Id:        uuid.New(),
		Username:  currentUsername,
		Type:      req.Type,
		Endpoint:  req.SubscriptionInfo.Endpoint,
		P256dh:    req.SubscriptionInfo.SubscriptionKeys.P256dh,
		Auth:      req.SubscriptionInfo.SubscriptionKeys.Auth,
		ExpoToken: req.Token,
	}

	err := service.pushSubscriptionRepo.CreatePushSubscription(&newPushSubscription)
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
		if pushSubscription.Type == "web" {
			go service.sendWebPushNotification(&pushSubscription, notificationString) // send push message in background
			continue
		}
		if pushSubscription.Type == "expo" {
			go service.sendExpoPushNotification(&pushSubscription, notificationString) // send push message in background
			continue
		}
	}
}

// sendWebPushNotification sends a push notification to a web client using the webpush-go library and provided keys
func (service *PushSubscriptionService) sendWebPushNotification(pushSubscription *models.PushSubscription, notificationString string) {
	// send notification to web client
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

	// Print response information
	fmt.Println("Notification sent to", pushSubscription.Username, "with type web")
	fmt.Println("Status:", resp.Status)
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println("Response Body:", string(bodyBytes))

	// If the subscription is deactivated or expired, delete it
	if resp.StatusCode == http.StatusGone {
		_ = service.pushSubscriptionRepo.DeletePushSubscriptionById(pushSubscription.Id.String())
		return
	}
}

// sendExpoPushNotification sends a push notification to an expo client using the expo API
func (service *PushSubscriptionService) sendExpoPushNotification(pushSubscription *models.PushSubscription, notificationString string) {
	// Create request
	expoApiUrl := "https://exp.host/--/api/v2/push/send"

	data := map[string]interface{}{
		"to":    pushSubscription.ExpoToken, // token is sent in ExponentPushToken[...] format
		"title": "Notification from Server Beta",
		"body":  notificationString,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	// Send request to expo API
	req, err := http.NewRequest("POST", expoApiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err, ", error sending notification to", pushSubscription.Username)
		return
	}

	// Print response information
	fmt.Println("Notification sent to", pushSubscription.Username, "with type expo")
	fmt.Println("Status:", resp.Status)
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println("Response Body:", string(bodyBytes))

	type Response struct {
		Data struct {
			Status  string `json:"status"`
			Message string `json:"message,omitempty"`
			Details struct {
				Error string `json:"error,omitempty"`
			} `json:"details,omitempty"`
		} `json:"data"`
	}

	var response Response
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		fmt.Println("Error response JSON:", err)
		return
	}

	// If the notification could not be sent, delete the subscription
	if response.Data.Details.Error == "DeviceNotRegistered" {
		_ = service.pushSubscriptionRepo.DeletePushSubscriptionById(pushSubscription.Id.String())
		return
	}
}
