package api

import (
	"crypto/subtle"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = time.Hour

func issueToken(secret, subject string) (string, time.Time, error) {
	exp := time.Now().Add(tokenTTL)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(exp),
	})
	s, err := tok.SignedString([]byte(secret))
	return s, exp, err
}

func loginHandler(cfg Config) fiber.Handler {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	return func(c *fiber.Ctx) error {
		var req request
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "cuerpo JSON invalido")
		}
		okUser := subtle.ConstantTimeCompare([]byte(req.Username), []byte(cfg.APIUser)) == 1
		okPass := subtle.ConstantTimeCompare([]byte(req.Password), []byte(cfg.APIPassword)) == 1
		if !okUser || !okPass {
			return fiber.NewError(fiber.StatusUnauthorized, "credenciales invalidas")
		}
		token, exp, err := issueToken(cfg.JWTSecret, req.Username)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "no se pudo emitir el token")
		}
		return c.JSON(fiber.Map{"token": token, "expiresAt": exp, "tokenType": "Bearer"})
	}
}

func requireJWT(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := strings.TrimPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
		if raw == "" || raw == c.Get(fiber.HeaderAuthorization) {
			return fiber.NewError(fiber.StatusUnauthorized, "falta el header Authorization: Bearer <token>")
		}
		_, err := jwt.Parse(raw, func(*jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "token invalido o expirado")
		}
		return c.Next()
	}
}
