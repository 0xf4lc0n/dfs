package dtos

import "time"

type ShareDto struct {
	FileId         uint      `json:"fileId"`
	UserId         uint      `json:"userId"`
	SharedById     uint      `json:"sharedById"`
	ExpirationTime time.Time `json:"expirationTime"`
}
