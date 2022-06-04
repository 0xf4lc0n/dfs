package dtos

import "time"

type SharedFileDto struct {
	UniqueName  string    `json:"uniqueName"`
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`
	SharedBy    string    `json:"sharedBy"`
	AvailableTo time.Time `json:"availableTo"`
}
