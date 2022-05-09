package routes

import (
	"dfs/storage/controllers"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Post("/api/file", controllers.UploadFile)
	app.Get("/api/file/:fileName", controllers.DownloadFile)
}
