package utils

import (
	"regexp"
)

// ExtractHashtags extracts all hashtags from a given text and returns them as a slice of strings
func ExtractHashtags(text string) []string {
	re := regexp.MustCompile(`#\w+`)
	matches := re.FindAllString(text, -1)

	hashtagsMap := make(map[string]bool) // use map first to remove duplicates
	for _, match := range matches {
		// Remove "#" from string and add to map
		hashtagsMap[match[1:]] = true
	}

	// Convert map keys to slice
	var hashtags []string
	for hashtag := range hashtagsMap {
		hashtags = append(hashtags, hashtag)
	}

	return hashtags
}
