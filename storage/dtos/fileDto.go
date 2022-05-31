package dtos

type FileDto struct {
	UniqueName string `json:"uniqueName"`
	Content    []byte `json:"content"`
}
