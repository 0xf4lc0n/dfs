package models

import "time"

type VerificationData struct {
	Id        uint      `json:"id"`
	Email     string    `json:"email" gorm:"unique"`
	Code      string    `json:"code" gorm:"unique"`
	ExpiresAt time.Time `json:"expiresAt"`
}
