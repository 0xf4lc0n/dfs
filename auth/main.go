package main

import (
	"dfs/auth/di"
	"dfs/auth/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	di.InitializeServices()
	defer di.Logger.Sync()

	go di.RpcService.RegisterValidateJwt()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	routes.Setup(app)

	app.Listen(":8080")
}
