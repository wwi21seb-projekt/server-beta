package utils_test

import (
	"encoding/base64"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateUsername tests the ValidateUsername function using multiple examples
func TestValidateUsername(t *testing.T) {
	validator := utils.NewValidator()

	testCases := []struct {
		username string
		expected bool
	}{
		{"ValidUser123", true},
		{"Valid-User_1.23", true},
		{"Invalid User", false},
		{"", false},
		{"NicknameWithEmoji😊", false},
		{"aVeryLongUsernameThatExceedsTwentyCharacters", false},
	}

	for _, tc := range testCases {
		result := validator.ValidateUsername(tc.username)
		if result != tc.expected {
			t.Errorf("ValidateUsername(%v): expected %v, got %v", tc.username, tc.expected, result)
		}
	}
}

// TestValidateNickname tests the ValidateNickname function using multiple examples
func TestValidateNickname(t *testing.T) {
	validator := utils.NewValidator()

	testCases := []struct {
		nickname string
		expected bool
	}{
		{"ValidNickname", true},
		{"Valid Nickname", true},
		{"", true},
		{"NicknameWithEmoji😊", true},
		{"aVeryLongNicknameThatExceedsTwentyFiveCharacters", false},
	}

	for _, tc := range testCases {
		result := validator.ValidateNickname(tc.nickname)
		if result != tc.expected {
			t.Errorf("ValidateNickname(%v): expected %v, got %v", tc.nickname, tc.expected, result)
		}
	}
}

// TestValidateEmail tests the ValidateEmailSyntax function using multiple examples
func TestValidateEmail(t *testing.T) {
	validator := utils.NewValidator()

	testCases := []struct {
		email    string
		expected bool
	}{
		{"valid.email@example.com", true},
		{"invalid-email", false},
		{"", false},
		{"another.valid-email@example.co.uk", true},
	}

	for _, tc := range testCases {
		result := validator.ValidateEmailSyntax(tc.email)
		if result != tc.expected {
			t.Errorf("ValidateEmailSyntax(%v): expected %v, got %v", tc.email, tc.expected, result)
		}
	}
}

// TestValidatePassword tests the ValidatePassword function using multiple examples
func TestValidatePassword(t *testing.T) {
	validator := utils.NewValidator()

	testCases := []struct {
		password string
		expected bool
	}{
		{"ValidPass123!", true},
		{"short", false},
		{"NoDigits123", false},
		{"nouppercase123!", false},
		{"NOLOWERCASE123!", false},
		{"NoSpecialChar123", false},
		{"Space   Space", false},
		{"Abcd123!", true},
		{"Abcd12!", false},
		{strings.Repeat("Aa2!", 8), true}, // 32 characters
		{strings.Repeat("A", 33), false},  // 33 characters
	}

	for _, tc := range testCases {
		result := validator.ValidatePassword(tc.password)
		if result != tc.expected {
			t.Errorf("ValidatePassword(%v): expected %v, got %v", tc.password, tc.expected, result)
		}
	}
}

// TestValidateStatus tests the ValidateStatus function using multiple examples
func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"Empty Status", "", true},
		{"Valid Status", "This is a valid status.", true},
		{"Exact Limit Status", string(make([]rune, 128)), true},
		{"Too Long Status", string(make([]rune, 129)), false},
	}

	validator := utils.NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validator.ValidateStatus(tt.status); got != tt.want {
				t.Errorf("ValidateStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateImage tests the ValidateImage function using multiple image examples from the tests/resources folder
func TestValidateImage(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{"Valid JPEG Image", "../../tests/resources/valid.jpeg", true},
		{"Valid WEBP Image", "../../tests/resources/valid.webp", true},
		{"Empty JPEG Image", "../../tests/resources/empty.jpeg", false},
		{"Empty WEBP Image", "../../tests/resources/empty.webp", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tt.filePath)
			imageData, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read image file: %s", err)
			}

			validator := utils.NewValidator()

			base64String := base64.StdEncoding.EncodeToString(imageData)
			if err != nil {
				log.Fatal("Error converting image to Base64: ", err)
			}
			isValid, _, err := validator.ValidateImage(base64String)

			if isValid != tt.want {
				t.Errorf("ValidateImage() = %v, want %v", isValid, tt.want)
			}
		})
	}
}

// TestValidateLatitude tests the ValidateLatitude function using multiple examples
func TestValidateLatitude(t *testing.T) {
	tests := []struct {
		name     string
		latitude float64
		want     bool
	}{
		{"Valid Positive", 90, true},
		{"Valid Negative", -90, true},
		{"Valid Zero", 0, true},
		{"Invalid Out of Range (Negative)", -90.1234, false},
		{"Invalid Out of Range (Positive)", 90.1234, false},
	}

	validator := utils.NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validator.ValidateLatitude(tt.latitude); got != tt.want {
				t.Errorf("ValidateLatitude() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateLongitude tests the ValidateLongitude function using multiple examples
func TestValidateLongitude(t *testing.T) {
	tests := []struct {
		name      string
		longitude float64
		want      bool
	}{
		{"Valid Positive", 180, true},
		{"Valid Negative", -180, true},
		{"Valid Zero", 0, true},
		{"Invalid Out of Range (Negative)", -180.1234567890, false},
		{"Invalid Out of Range (Positive)", 180.1234567890, false},
	}

	validator := utils.NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validator.ValidateLongitude(tt.longitude); got != tt.want {
				t.Errorf("ValidateLongitude() = %v, want %v", got, tt.want)
			}
		})
	}
}
