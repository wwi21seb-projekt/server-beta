package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a given password with a random salt and returns hash
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	return string(hash), err
}

// CheckPassword checks if a given password matches with the given hash (contains salt)
func CheckPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
