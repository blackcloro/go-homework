package main

import (
	"regexp"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// middlewareSetup returns a slice of middleware functions for setup
func middlewareSetup() []any {
	return []any{
		// Set the Cache-Control header for all responses
		func(c fiber.Ctx) error {
			c.Set("Cache-Control", "max-age=0, private, must-revalidate")
			return c.Next()
		},
		recover.New(),
		requestid.New(),
		logger.New(),
	}
}

// customErrorHandler is a custom error handler for the application
func customErrorHandler(ctx fiber.Ctx, _ error) error {
	return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"errors": fiber.Map{
			"detail": "Not Found",
		},
	})
}

// methodValidationMiddleware validates the HTTP method for each route
func methodValidationMiddleware(expectedMethods map[*regexp.Regexp][]string) fiber.Handler {
	return func(c fiber.Ctx) error {
		path := c.Path()
		for pattern, methods := range expectedMethods {
			if pattern.MatchString(path) {
				for _, method := range methods {
					if c.Method() == method {
						return c.Next()
					}
				}
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"errors": fiber.Map{
						"detail": "Not Found",
					},
				})
			}
		}
		return c.Next()
	}
}
