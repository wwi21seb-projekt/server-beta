package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "your_password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword failed: %v", err)
	}

	if len(hash) == 0 {
		t.Errorf("Generated hash is empty")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "your_password"
	hash, _ := HashPassword(password)

	if !CheckPassword(password, hash) {
		t.Errorf("CheckPassword failed for correct password")
	}

	if CheckPassword("wrong_password", hash) {
		t.Errorf("CheckPassword should fail for wrong password")
	}
}
