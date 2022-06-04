package dtos

import "time"

type SharedForDto struct {
	UniqueName  string    `json:"uniqueName"`
	Name        string    `json:"name"`
	SharedFor   []string  `json:"sharedFor"`
	AvailableTo time.Time `json:"availableToTo"`
}
