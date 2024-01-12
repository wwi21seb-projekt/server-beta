package utils

import (
	"regexp"
)

// ExtractHashtags extracts all hashtags from a given text and returns them as a slice of strings
func ExtractHashtags(text string) []string {
	re := regexp.MustCompile(`#\w+`)
	matches := re.FindAllString(text, -1)

	hashtagsMap := make(map[string]bool)
	var hashtags []string

	for _, match := range matches {
		// Remove "#" from string
		hashtag := match[1:]

		// Add to slice if not exist in map
		if !hashtagsMap[hashtag] {
			hashtagsMap[hashtag] = true
			hashtags = append(hashtags, hashtag)
		}
	}

	return hashtags
}
