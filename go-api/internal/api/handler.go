package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jinfanzon/reto-tecnico/go-api/internal/matrix"
)

type StatsProvider interface {
	Fetch(q, r matrix.Matrix, authHeader string) (*matrix.Statistics, error)
}

func qrHandler(stats StatsProvider) fiber.Handler {
	type request struct {
		Matrix matrix.Matrix `json:"matrix"`
	}
	return func(c *fiber.Ctx) error {
		var req request
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "cuerpo JSON invalido; se espera {\"matrix\": [[...]]}")
		}
		q, r, err := req.Matrix.QR()
		if err != nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
		}
		s, err := stats.Fetch(q, r, c.Get(fiber.HeaderAuthorization))
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		return c.JSON(fiber.Map{
			"rows": len(req.Matrix), "columns": len(req.Matrix[0]),
			"q": q, "r": r, "statistics": s,
		})
	}
}
