package utils

import (
	"os"
	"strings"
)

// CensorEmail censors the email address for the response
// converts testmail@gmail.com to tes*****@gmail.com
func CensorEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	name := parts[0]
	if len(name) > 3 {
		name = name[:3] + strings.Repeat("*", len(name)-3)
	}
	return name + "@" + parts[1]
}

// FormatImageUrl formats the image url for the response
// adds the base url with image id and file extension
func FormatImageUrl(imageId string, extension string) string {
	return os.Getenv("SERVER_URL") + "/api/images/" + imageId + "." + extension
}
