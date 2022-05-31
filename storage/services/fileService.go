package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"dfs/storage/config"
	"encoding/base64"
	"go.uber.org/zap"
	"io"
	"os"
	"path"
)

type FileService struct {
	config *config.Config
	logger *zap.Logger
}

func NewFileService(cfg *config.Config, logger *zap.Logger) *FileService {
	return &FileService{config: cfg, logger: logger}
}

func (fs *FileService) EncryptAndSaveFile(filePath string, fileContent []byte, key []byte) bool {
	encrypted := fs.Encrypt(fileContent, key)
	encoded := fs.Encode(encrypted)
	return fs.SaveFileOnDisk(filePath, encoded)
}

func (fs *FileService) DecryptAndReadFileContent(filePath string, key []byte) []byte {
	encoded := fs.ReadFileFromDisk(filePath)

	if encoded == nil {
		return nil
	}

	encrypted := fs.Decode(encoded)
	return fs.Decrypt(encrypted, key)
}

func (fs *FileService) SaveFileOnDisk(filePath string, fileContent []byte) bool {
	savePath := path.Join(fs.config.FileStoragePath, filePath)

	if err := os.WriteFile(savePath, fileContent, 0644); err != nil {
		fs.logger.Error("Cannot save file on the disk", zap.Error(err))
		return false
	}

	return true
}

func (fs *FileService) ReadFileFromDisk(filePath string) []byte {
	readPath := path.Join(fs.config.FileStoragePath, filePath)
	fileContent, err := os.ReadFile(readPath)

	if err != nil {
		fs.logger.Error("Cannot read file", zap.Error(err))
		return nil
	}

	return fileContent
}

func (fs *FileService) RemoveFileFromDisk(filePath string) bool {
	deletePath := path.Join(fs.config.FileStoragePath, filePath)

	if err := os.Remove(deletePath); err != nil {
		fs.logger.Error("Cannot remove file from the disk", zap.Error(err))
		return false
	}

	return true
}

func (fs *FileService) Encode(plainText []byte) []byte {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(plainText)))
	base64.StdEncoding.Encode(encoded, plainText)
	return encoded
}

func (fs *FileService) Decode(encoded []byte) []byte {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
	n, err := base64.StdEncoding.Decode(decoded, encoded)

	if err != nil {
		fs.logger.Panic("Cannot decode data from Base64", zap.Error(err))
	}

	decoded = decoded[:n]

	return decoded
}

func (fs *FileService) Encrypt(plainText []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)

	if err != nil {
		fs.logger.Panic("Cannot create cipher", zap.Error(err))
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))

	iv := cipherText[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		fs.logger.Panic("Cannot fill iv with random bytes", zap.Error(err))
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return cipherText
}

func (fs *FileService) Decrypt(cipherText []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)

	if err != nil {
		fs.logger.Panic("Cannot create cipher", zap.Error(err))
	}

	if len(cipherText) < aes.BlockSize {
		fs.logger.Panic("Cipher text is too short")
	}

	iv := cipherText[:aes.BlockSize]

	cipherText = cipherText[aes.BlockSize:]
	plainText := make([]byte, len(cipherText))

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(plainText, cipherText)

	return plainText
}

func (fs *FileService) CreateDirectory(directoryName string) bool {
	directoryPath := path.Join(fs.config.FileStoragePath, directoryName)

	if err := os.Mkdir(directoryPath, 0755); err != nil {
		fs.logger.Error("Cannot create home directory", zap.Error(err))
		return false
	}

	return true
}
