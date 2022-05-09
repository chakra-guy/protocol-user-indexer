package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	log.Fatal().Err(app.Listen(":3000"))
}
