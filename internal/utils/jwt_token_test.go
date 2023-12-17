package utils_test

import (
	"github.com/marcbudd/server-beta/internal/utils"
	"testing"
)

func TestGenerateAccessToken(t *testing.T) {
	username := "testUser"
	token, err := utils.GenerateAccessToken(username)
	if err != nil {
		t.Errorf("Error generating access token: %v", err)
	}

	if token == "" {
		t.Errorf("Generated token is empty")
	}
}

func TestVerifyAccessToken(t *testing.T) {
	username := "testUser"
	token, err := utils.GenerateAccessToken(username)
	if err != nil {
		t.Errorf("Error generating access token: %v", err)
	}

	// Test valid token
	returnedUsername, err := utils.VerifyAccessToken(token)
	if err != nil || returnedUsername != username {
		t.Errorf("Error verifying valid token: %v", err)
	}

	// Test invalid token
	_, err = utils.VerifyAccessToken("invalid.token")
	if err == nil {
		t.Error("Expected error verifying invalid token, got nil")
	}
}
