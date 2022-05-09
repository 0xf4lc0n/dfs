package controllers

import (
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const STORAGE_PATH = `C:\Users\Falcon\Desktop\Files`

func UploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	fileName := uuid.New().String()
	fullFilePath := path.Join(STORAGE_PATH, fileName)
	return c.SaveFile(file, fullFilePath)
}

func DownloadFile(c *fiber.Ctx) error {
	fileName := c.Params("fileName")
	fullFilePath := path.Join(STORAGE_PATH, fileName)
	return c.Download(fullFilePath)
}
