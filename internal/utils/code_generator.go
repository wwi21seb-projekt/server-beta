package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateSixDigitCode generates a code that is used for email verification
func GenerateSixDigitCode() (int64, error) {
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return n.Int64() + 100000, nil
}
