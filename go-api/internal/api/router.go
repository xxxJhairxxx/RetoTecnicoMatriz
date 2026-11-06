package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Config struct {
	JWTSecret   string
	APIUser     string
	APIPassword string
}

func NewRouter(cfg Config, stats StatsProvider) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: errorHandler})
	app.Use(recover.New(), logger.New(), cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Post("/api/v1/auth/token", loginHandler(cfg))

	protected := app.Group("/api/v1", requireJWT(cfg.JWTSecret))
	protected.Post("/qr", qrHandler(stats))

	return app
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
	}
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}
