package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/truemail-rb/truemail-go"
	"golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"regexp"
	"strings"
	"unicode"
)

type ValidatorInterface interface {
	ValidateUsername(username string) bool
	ValidateNickname(nickname string) bool
	ValidateEmailSyntax(email string) bool
	ValidateEmailExistance(email string) bool
	ValidatePassword(password string) bool
	ValidateStatus(status string) bool
	ValidateImage(imageData string) (bool, string, error)
	ValidateLatitude(latitude float64) bool
	ValidateLongitude(longitude float64) bool
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
	return len(status) <= 128
}

// ValidateImage validates if an image is correct file type and can be decoded, additionally returns file format
func (v *Validator) ValidateImage(base64String string) (bool, string, error) {
	// Decode the Base64 string
	imageData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return false, "", err
	}

	// Detect image type by checking magic numbers
	var contentType string
	if bytes.HasPrefix(imageData, []byte{0xff, 0xd8, 0xff}) {
		contentType = "image/jpeg"
	} else if bytes.HasPrefix(imageData, []byte{'R', 'I', 'F', 'F'}) && bytes.Contains(imageData[:16], []byte{'W', 'E', 'B', 'P'}) {
		contentType = "image/webp"
	} else if bytes.HasPrefix(imageData, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		contentType = "image/png"
	} else if bytes.HasPrefix(imageData, []byte{'G', 'I', 'F'}) {
		contentType = "image/gif"
	} else if bytes.HasPrefix(imageData, []byte{'<', 's', 'v', 'g'}) || bytes.HasPrefix(imageData, []byte{'<', '?', 'x', 'm', 'l'}) {
		contentType = "image/svg+xml"
	} else {
		return false, "", errors.New("unknown image format")
	}

	// Then check if image can be decoded if it's not SVG
	if contentType != "image/svg+xml" {
		var err error
		if contentType == "image/webp" {
			_, err = webp.Decode(bytes.NewReader(imageData))
		} else {
			_, _, err = image.Decode(bytes.NewReader(imageData))
		}
		if err != nil {
			return false, "", err
		}
	} else {
		// Additional SVG validation can be added here if needed
		if !strings.Contains(string(imageData), "<svg") {
			return false, "", errors.New("invalid SVG content")
		}
	}

	return true, contentType, nil
}

// ValidateLatitude validates if latitude is in valid range
func (v *Validator) ValidateLatitude(latitude float64) bool {
	return latitude >= -90 && latitude <= 90
}

// ValidateLongitude validates if longitude is in valid range
func (v *Validator) ValidateLongitude(longitude float64) bool {
	return longitude >= -180 && longitude <= 180
}
