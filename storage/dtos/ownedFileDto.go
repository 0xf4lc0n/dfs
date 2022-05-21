package dtos

type OwnedFileDto struct {
	FileId  uint `json:"fileId"`
	OwnerId uint `json:"ownerId"`
}
