package dtos

type ReadFileDto struct {
	ReadPath      string `json:"savePath"`
	DecryptionKey []byte `json:"decryptionKey"`
}
