package utils_test

import (
	"github.com/marcbudd/server-beta/internal/utils"
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
	}

	for _, tc := range testCases {
		result := validator.ValidatePassword(tc.password)
		if result != tc.expected {
			t.Errorf("ValidatePassword(%v): expected %v, got %v", tc.password, tc.expected, result)
		}
	}
}
