package utils

import (
	"reflect"
	"testing"
)

func TestExtractHashtags(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "Standard hashtags",
			text:     "My #first Post here on #Social_Media",
			expected: []string{"first", "Social_Media"},
		},
		{
			name:     "No hashtags",
			text:     "A post without hashtags",
			expected: []string{},
		},
		{
			name:     "Mixed hashtags",
			text:     "#Hello, this is a #1Test_1Post with #123Numbers",
			expected: []string{"Hello", "1Test_1Post", "123Numbers"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractHashtags(tt.text)

			// Test if both slices are empty
			if len(result) == 0 && len(tt.expected) == 0 {
				return
			}

			// Test if both slices have the same elements
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExtractHashtags(%s) = %v, expected %v", tt.text, result, tt.expected)
			}
		})
	}
}
