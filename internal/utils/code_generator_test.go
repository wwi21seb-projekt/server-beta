package utils

import "testing"

// TestGenerateSixDigitCode tests if GenerateSixDigitCode returns a six digit code
func TestGenerateSixDigitCode(t *testing.T) {
	for i := 0; i < 1000; i++ {
		code, err := GenerateSixDigitCode()
		if err != nil {
			t.Errorf("GenerateSixDigitCode() returned an error: %v", err)
		}
		if code < 100000 || code > 999999 {
			t.Errorf("GenerateSixDigitCode() returned an invalid code: %d", code)
		}
	}
}