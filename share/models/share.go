package models

import "time"

type Share struct {
	FileId         uint      `json:"fileId" gorm:"primaryKey;autoIncrement:false"`
	UserId         uint      `json:"userId" gorm:"primaryKey;autoIncrement:false"`
	SharedById     uint      `json:"sharedById"`
	ExpirationTime time.Time `json:"expirationTime"`
}
