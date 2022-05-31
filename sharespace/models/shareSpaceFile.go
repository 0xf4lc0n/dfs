package models

import "time"

type ShareSpaceFile struct {
	Id           uint      `json:"id"`
	ShareSpaceId uint      `json:"shareSpaceId"`
	UniqueName   string    `json:"uniqueName" gorm:"unique"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	CreationDate time.Time `json:"creationDate"`
	OwnerId      uint      `json:"ownerId"`
}
