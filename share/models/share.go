package models

import "time"

type Share struct {
	FileId         uint      `json:"fileId" gorm:"primaryKey;autoIncrement:false"`
	SharedForId    uint      `json:"sharedForId" gorm:"primaryKey;autoIncrement:false"`
	SharedById     uint      `json:"sharedById"`
	ExpirationTime time.Time `json:"expirationTime"`
}
