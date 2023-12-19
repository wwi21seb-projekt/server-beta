package utils

import (
	"regexp"
)

// ExtractHashtags extracts all hashtags from a given text and returns them as a slice of strings
func ExtractHashtags(text string) []string {
	// Regex für Hashtags: beings with # and can contain letters, numbers and underscores
	re := regexp.MustCompile(`#\w+`)
	matches := re.FindAllString(text, -1)

	var hashtags []string
	for _, match := range matches {
		// Remove "#" from string
		hashtags = append(hashtags, match[1:])
	}

	return hashtags
}
