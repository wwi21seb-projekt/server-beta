package utils_test

import (
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
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

// TestValidateCoordinate tests the ValidateCoordinate function using multiple examples
func TestValidateCoordinate(t *testing.T) {
	tests := []struct {
		name       string
		coordinate string
		want       bool
	}{
		{"Valid Positive", "123,4567890", true},
		{"Valid Negative", "-123,4567", true},
		{"Valid Zero", "0,0", true},
		{"Valid No Decimal", "123", true},
		{"Valid Max Decimal", "123,1234567890", true},
		{"Invalid Empty", "", false},
		{"Invalid Too Long", "123,12345678901", false},
		{"Invalid Characters", "abc,def", false},
		{"Invalid Injection", "123; DROP TABLE users;", false},
		{"Invalid Only Decimal", ",1234", false},
		{"Invalid Multiple Commas", "123,456,789", false},
		{"Invalid Negative Decimal", "123,-4567", false},
		{"Invalid Out of Range", "1234,5678", false},
	}

	validator := utils.NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validator.ValidateCoordinate(tt.coordinate); got != tt.want {
				t.Errorf("ValidateStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateLatitude tests the ValidateLatitude function using multiple examples
func TestValidateLatitude(t *testing.T) {
	tests := []struct {
		name     string
		latitude string
		want     bool
	}{
		{"Valid Positive", "90", true},
		{"Valid Negative", "-90", true},
		{"Valid Zero", "0", true},
		{"Valid Max Decimal", "90.1234567890", false},
		{"Invalid Empty", "", false},
		{"Invalid Too Long", "90.12345678901", false},
		{"Invalid Characters", "abc", false},
		{"Invalid Injection", "90; DROP TABLE users;", false},
		{"Invalid Multiple Dots", "90.123.456", false},
		{"Invalid Out of Range", "90.1234", false},
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
		longitude string
		want      bool
	}{
		{"Valid Positive", "180", true},
		{"Valid Negative", "-180", true},
		{"Valid Zero", "0", true},
		{"Valid Max Decimal", "180.1234567890", false},
		{"Invalid Empty", "", false},
		{"Invalid Too Long", "180.12345678901", false},
		{"Invalid Characters", "abc", false},
		{"Invalid Injection", "180; DROP TABLE users;", false},
		{"Invalid Multiple Dots", "180.123.456", false},
		{"Invalid Out of Range", "180.1234", false},
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
