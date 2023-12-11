package models

import "time"

type VerificationCode struct {
	Username       string    `gorm:"primary_key"`
	Code           string    `gorm:"primary_key;column:code;not_null"`
	ExpirationTime time.Time `gorm:"column:expiration_time;not_null"`
}
