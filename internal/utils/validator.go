package utils

import (
	"bytes"
	"github.com/truemail-rb/truemail-go"
	_ "golang.org/x/image/webp" // Needs to be imported for webp decoding, but is used implicitly
	"image"
	"os"
	"regexp"
	"strings"
	"strconv"
	"unicode"
)

type ValidatorInterface interface {
	ValidateUsername(username string) bool
	ValidateNickname(nickname string) bool
	ValidateEmailSyntax(email string) bool
	ValidateEmailExistance(email string) bool
	ValidatePassword(password string) bool
	ValidateStatus(status string) bool
	ValidateImage(imageData []byte, contentType string) bool
	ValidateCoordinate(coordinate string) bool
	ValidateLatitude(latitude string) bool
	ValidateLongitude(longitude string) bool
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
func (v *Validator) ValidateImage(imageData []byte, contentType string) bool {
	// First check magic numbers of image data
	if contentType == "image/jpeg" {
		if !bytes.HasPrefix(imageData, []byte{0xff, 0xd8, 0xff}) {
			return false
		}
	} else if contentType == "image/webp" {
		if !bytes.HasPrefix(imageData, []byte{'R', 'I', 'F', 'F'}) || !bytes.Contains(imageData[:16], []byte{'W', 'E', 'B', 'P'}) {
			return false
		}
	} else {
		return false
	}

	// Then check if image can be decoded
	_, fileFormat, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return false
	}

	// Check if file format matches content type
	if !strings.Contains(contentType, fileFormat) {
		return false
	}

	return true
}

// ValidateCoordinate validates if a string matches the coordinate format
func (v *Validator) ValidateCoordinate(coordinate string) bool {
	re := regexp.MustCompile(`^-?\d{1,3}.\d{0,10}$`)
	return re.MatchString(coordinate)
}

// ValidateLatitude validates if a string is a valid latitude
func (v *Validator) ValidateLatitude(latitude string) bool {
	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		return false
	}
	return lat >= -90 && lat <= 90 // latitude must be between -90 and 90
}

// ValidateLongitude validates if a string is a valid longitude
func (v *Validator) ValidateLongitude(longitude string) bool {
	lon, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		return false
	}
	return lon >= -180 && lon <= 180 // longitude must be between -180 and 180
}
