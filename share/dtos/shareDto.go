package dtos

import "time"

type ShareDto struct {
	FileId         uint      `json:"fileId"`
	SharedToId     uint      `json:"sharedToId"`
	SharedById     uint      `json:"sharedById"`
	ExpirationTime time.Time `json:"expirationTime"`
}
