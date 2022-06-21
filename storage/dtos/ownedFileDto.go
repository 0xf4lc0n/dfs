package dtos

type OwnedFileDto struct {
	FileUniqueName string `json:"fileUniqueName"`
	OwnerId        uint64 `json:"ownerId"`
}
