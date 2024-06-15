package utils

import (
	"bytes"
	"encoding/xml"
	"github.com/truemail-rb/truemail-go"
	"golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"regexp"
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
	ValidateLatitude(latitude float64) bool
	ValidateLongitude(longitude float64) bool
	ValidateImage(imageData []byte) (bool, string, int, int)
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

// ValidateLatitude validates if latitude is in valid range
func (v *Validator) ValidateLatitude(latitude float64) bool {
	return latitude >= -90 && latitude <= 90
}

// ValidateLongitude validates if longitude is in valid range
func (v *Validator) ValidateLongitude(longitude float64) bool {
	return longitude >= -180 && longitude <= 180
}

// ValidateImage validates if an image is correct file type and can be "opened", additionally returns file format, width and height
func (v *Validator) ValidateImage(imageData []byte) (bool, string, int, int) {
	// If file size is larger than 10 MB, return false
	if len(imageData) > 10*1024*1024 { // 10 MB
		return false, "", 0, 0
	}

	// Decode to validate image: png, jpeg
	img, format, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err == nil && (format == "png" || format == "jpeg") {
		return true, format, img.Width, img.Height
	}

	// Decode to validate image: webp
	img, err = webp.DecodeConfig(bytes.NewReader(imageData))
	if err == nil {
		return true, "webp", img.Width, img.Height
	}

	// Decode to validate image: svg
	if bytes.HasPrefix(imageData, []byte("<svg")) && bytes.HasSuffix(imageData, []byte("</svg>")) {
		// Check for disallowed elements or attributes
		if bytes.Contains(imageData, []byte("<script")) || // Check for <script> tags
			bytes.Contains(imageData, []byte("onload=")) || // Check for onload event
			bytes.Contains(imageData, []byte("onclick=")) || // Check for onclick event
			bytes.Contains(imageData, []byte("onmouseover=")) || // Check for onmouseover event
			bytes.Contains(imageData, []byte("data:")) { // Check for data URIs
			return false, "", 0, 0
		}

		// Decode svg to get width and height
		type SVG struct {
			XMLName xml.Name `xml:"svg"`
			Width   string   `xml:"width,attr"`
			Height  string   `xml:"height,attr"`
		}
		var svg SVG
		decoder := xml.NewDecoder(bytes.NewReader(imageData))
		if err := decoder.Decode(&svg); err != nil {
			return false, "", 0, 0
		}
		// Convert width and height to integers
		// If conversion return 0, return false
		width, _ := strconv.Atoi(svg.Width)
		height, _ := strconv.Atoi(svg.Height)
		if width == 0 || height == 0 {
			return false, "", 0, 0
		}

		return true, "svg", width, height
	}

	// If none of the above conditions are met, return false
	return false, "", 0, 0
}
