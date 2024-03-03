package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	port := getPort()

	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	app.Use(middlewareSetup()...)

	setupRoutes(app)

	log.Fatal(app.Listen(port))
}
