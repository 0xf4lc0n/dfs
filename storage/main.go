package main

import (
	"dfs/shared/rpc"
	"dfs/storage/di"
	"dfs/storage/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	di.InitializeServices()
	defer di.Logger.Sync()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	app.Use(func(c *fiber.Ctx) error {
		cookie := c.Cookies("jwt")

		if rpc.IsAuthenticated(cookie, di.Logger) == false {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.Next()
	})

	routes.Setup(app)

	app.Listen(":8081")
}
