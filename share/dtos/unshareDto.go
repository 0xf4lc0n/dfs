package dtos

type UnshareDto struct {
	FileId      uint `json:"fileId"`
	SharedForId uint `json:"sharedForId"`
}
