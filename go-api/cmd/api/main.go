package main

import (
	"log"
	"os"

	"github.com/jinfanzon/reto-tecnico/go-api/internal/api"
	"github.com/jinfanzon/reto-tecnico/go-api/internal/stats"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	apiPassword := os.Getenv("API_PASSWORD")
	if jwtSecret == "" || apiPassword == "" {
		log.Fatal("JWT_SECRET y API_PASSWORD son obligatorios")
	}

	cfg := api.Config{
		JWTSecret:   jwtSecret,
		APIUser:     env("API_USER", "demo"),
		APIPassword: apiPassword,
	}

	app := api.NewRouter(cfg, stats.NewClient(env("NODE_API_URL", "http://localhost:3000")))
	app.Static("/", env("PUBLIC_DIR", "./public"))

	log.Fatal(app.Listen(":" + env("PORT", "8080")))
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
