package utils_test

import (
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"testing"
)

// TestGenerateAccessToken tests the GenerateAccessToken function if it returns a token
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

// TestGenerateRefreshToken tests the GenerateRefreshToken function if it returns a token
func TestGenerateRefreshToken(t *testing.T) {
	username := "testUser"
	token, err := utils.GenerateRefreshToken(username)
	if err != nil {
		t.Errorf("Error generating access token: %v", err)
	}

	if token == "" {
		t.Errorf("Generated token is empty")
	}
}

// TestVerifyJWTTokenAccess tests the VerifyJWTToken function if it returns the correct username and false if the token is valid access token
func TestVerifyJWTTokenAccess(t *testing.T) {
	username := "testUser"
	token, err := utils.GenerateAccessToken(username)
	if err != nil {
		t.Errorf("Error generating access token: %v", err)
	}

	// Test valid token
	returnedUsername, isRefresh, err := utils.VerifyJWTToken(token)
	if err != nil || returnedUsername != username || isRefresh {
		t.Errorf("Error verifying valid token: %v", err)
	}

	// Test invalid token
	_, _, err = utils.VerifyJWTToken("invalid.token")
	if err == nil {
		t.Error("Expected error verifying invalid token, got nil")
	}
}

// TestVerifyJWTTokenRefresh tests the VerifyJWTToken function if it returns the correct username and true if the token is valid refresh token
func TestVerifyJWTTokenRefresh(t *testing.T) {
	username := "testUser"
	token, err := utils.GenerateRefreshToken(username)
	if err != nil {
		t.Errorf("Error generating access token: %v", err)
	}

	// Test valid token
	returnedUsername, isRefresh, err := utils.VerifyJWTToken(token)
	if err != nil || returnedUsername != username || !isRefresh {
		t.Errorf("Error verifying valid token: %v", err)
	}

	// Test invalid token
	_, _, err = utils.VerifyJWTToken("invalid.token")
	if err == nil {
		t.Error("Expected error verifying invalid token, got nil")
	}
}
