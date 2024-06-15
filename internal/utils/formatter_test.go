package utils_test

import (
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"os"
	"testing"
)

// TestCensorEmail tests the CensorEmail function
func TestCensorEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"testmail@gmail.com", "tes*****@gmail.com"},
		{"short@mail.com", "sho**@mail.com"},
		{"a@b.com", "a@b.com"},
		{"ab@c.com", "ab@c.com"},
		{"abc@def.com", "abc@def.com"},
		{"", ""},
		{"invalidEmail", "invalidEmail"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := utils.CensorEmail(tt.email)
			if result != tt.expected {
				t.Errorf("CensorEmail(%q) = %q; want %q", tt.email, result, tt.expected)
			}
		})
	}
}

// TestFormatImageUrl tests the FormatImageUrl function if it formats the image url correctly
func TestFormatImageUrl(t *testing.T) {
	err := os.Setenv("SERVER_URL", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name        string
		imageId     string
		extension   string
		expectedUrl string
	}{
		{
			name:        "Basic test case",
			imageId:     "123456789",
			extension:   "jpg",
			expectedUrl: "https://example.com/api/images/123456789.jpg",
		},
		{
			name:        "Different extension",
			imageId:     "987654321",
			extension:   "png",
			expectedUrl: "https://example.com/api/images/987654321.png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.FormatImageUrl(tc.imageId, tc.extension)
			if result != tc.expectedUrl {
				t.Errorf("FormatImageUrl returned unexpected result for %s, got: %s, want: %s", tc.name, result, tc.expectedUrl)
			}
		})
	}
}
