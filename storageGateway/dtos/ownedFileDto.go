package dtos

type OwnedFileDto struct {
	FileId  uint64 `json:"fileId"`
	OwnerId uint64 `json:"ownerId"`
}
