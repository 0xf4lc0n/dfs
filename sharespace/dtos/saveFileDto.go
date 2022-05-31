package dtos

type SaveFileDto struct {
	SavePath      string `json:"savePath"`
	Content       []byte `json:"content"`
	EncryptionKey []byte `json:"encryptionKey"`
}
