package utils

import (
	"github.com/truemail-rb/truemail-go"
	"os"
	"regexp"
	"unicode"
)

type ValidatorInterface interface {
	ValidateUsername(username string) bool
	ValidateNickname(nickname string) bool
	ValidateEmailSyntax(email string) bool
	ValidateEmailExistance(email string) bool
	ValidatePassword(password string) bool
	ValidateStatus(status string) bool
}

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

// ValidateUsername validates if a password meets specifications
func (v *Validator) ValidateUsername(username string) bool {
	if len(username) > 20 {
		return false
	}
	usernameRegex := `^[A-Za-z0-9_\-\.]+$`
	match, _ := regexp.MatchString(usernameRegex, username)
	return match
}

// ValidateNickname validates if a nickname meets specifications
func (v *Validator) ValidateNickname(nickname string) bool {
	return len(nickname) <= 25
}

// ValidateEmailSyntax validates if an email meets specifications
func (v *Validator) ValidateEmailSyntax(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	if match == false || len(email) > 128 {
		return false
	}
	return true
}

// ValidateEmailExistance tries to reach out to email to see if it really exists
func (v *Validator) ValidateEmailExistance(email string) bool {
	var configuration, _ = truemail.NewConfiguration(truemail.ConfigurationAttr{
		VerifierEmail:         os.Getenv("EMAIL_ADDRESS"),
		ValidationTypeDefault: "mx",
		SmtpFailFast:          true,
	})

	return truemail.IsValid(email, configuration)
}

// ValidatePassword validates if a password meets specifications
func (v *Validator) ValidatePassword(password string) bool {
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

// ValidateStatus validates if a status meets specifications
func (v *Validator) ValidateStatus(status string) bool {
	return len(status) <= 125
}
