package utils_test

import (
	"github.com/marcbudd/server-beta/internal/utils"
	"testing"
)

// TestHashPassword tests if HashPassword generates a hash to a given password
func TestHashPassword(t *testing.T) {
	password := "your_password"
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword failed: %v", err)
	}

	if len(hash) == 0 {
		t.Errorf("Generated hash is empty")
	}
}

// TestCheckPassword tests if CheckPassword succeeds for the right password and fails for the wrong one
func TestCheckPassword(t *testing.T) {
	password := "your_password"
	hash, _ := utils.HashPassword(password)

	if !utils.CheckPassword(password, hash) {
		t.Errorf("CheckPassword failed for correct password")
	}

	if utils.CheckPassword("wrong_password", hash) {
		t.Errorf("CheckPassword should fail for wrong password")
	}
}
