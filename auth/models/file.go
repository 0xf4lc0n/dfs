package models

import "time"

type File struct {
	Id           uint      `json:"id"`
	UniqueName   string    `json:"uniqueName" gorm:"unique"`
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creationDate"`
	OwnerId      uint      `json:"ownerId"`
}
