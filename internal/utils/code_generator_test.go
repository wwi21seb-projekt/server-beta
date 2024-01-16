package utils_test

import (
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"strconv"
	"testing"
	"unicode"
)

// TestGenerateSixDigitCode tests if GenerateSixDigitCode returns a six digit code
func TestGenerateSixDigitCode(t *testing.T) {
	code, err := utils.GenerateSixDigitCode()
	if err != nil {
		t.Fatalf("GenerateSixDigitCode() returned an error: %v", err)
	}

	codeStr := strconv.FormatInt(code, 10)
	if len(codeStr) != 6 {
		t.Errorf("Expected a 6-digit code, got %d digits", len(codeStr))
	}

	for _, r := range codeStr {
		if !unicode.IsDigit(r) {
			t.Errorf("Expected only digits in the code, found non-digit: %c", r)
		}
	}
}
