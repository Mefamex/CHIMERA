// file path: internal/server/apiGates.go
package server

import (
	"CHIMERA/scripts"

	"github.com/gofiber/fiber/v2"
)

func setupApiStackts(app *fiber.App) {
	app.Get("/api/stacks", func(c *fiber.Ctx) error {
		status, err := scripts.Stacks_info()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(status)
	})
}
