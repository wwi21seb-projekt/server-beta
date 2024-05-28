package utils_test

import (
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
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
		{"invalidemail", "invalidemail"},
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
