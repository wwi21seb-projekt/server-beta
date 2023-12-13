package utils

import (
	"regexp"
	"unicode"
)

// ValidateUsername validates if a password meets specifications
func ValidateUsername(username string) bool {
	if len(username) > 20 {
		return false
	}
	usernameRegex := `^[A-Za-z0-9_\-\.]+$`
	match, _ := regexp.MatchString(usernameRegex, username)
	return match
}

// ValidateNickname validates if a nickname meets specifications
func ValidateNickname(nickname string) bool {
	return len(nickname) <= 25
}

// ValidateEmail validates if an email meets specifications
func ValidateEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	if match == false || len(email) > 128 {
		return false
	}
	return true
}

// ValidatePassword validates if a password meets specifications
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	if len(password) > 32 {
		return false
	}
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasNumber = true
		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
			// No letter, no digit -> special character
			hasSpecial = true
		}

	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}
