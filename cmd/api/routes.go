package main

import (
	"regexp"

	"awesomeProject/pkg/api"
	"github.com/gofiber/fiber/v3"
)

func setupRoutes(app *fiber.App) {
	// Define expected methods for each path pattern
	expectedMethods := map[*regexp.Regexp][]string{
		regexp.MustCompile(`^/api/products$`):                    {"GET"},
		regexp.MustCompile(`^/api/orders$`):                      {"POST"},
		regexp.MustCompile(`^/api/orders/[^/]+$`):                {"GET", "PATCH"},
		regexp.MustCompile(`^/api/orders/[^/]+/products$`):       {"GET", "POST"},
		regexp.MustCompile(`^/api/orders/[^/]+/products/[^/]+$`): {"PATCH"},
	}

	// Custom method error handler middleware
	app.Use(methodValidationMiddleware(expectedMethods))

	// Endpoint definitions
	app.Get("/api/products", api.GetProducts)
	app.Post("/api/orders", api.CreateOrder)
	app.Get("/api/orders/:order_id", api.GetOrder)
	app.Patch("/api/orders/:order_id", api.UpdateOrderStatus)
	app.Post("/api/orders/:order_id/products", api.AddProductsToOrder)
	app.Get("/api/orders/:order_id/products", api.GetOrderProducts)
	app.Patch("/api/orders/:order_id/products/:product_id", api.ProductPatchHandler)
}
