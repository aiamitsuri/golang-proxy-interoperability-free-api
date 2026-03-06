package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

const (
	rustSource = "https://interoperability.onrender.com/filter"
	targetHost = "interoperability.onrender.com"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "Interoperability Proxy API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip} | ${status} | ${latency} | ${method} ${path}\n",
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "Slow down!"})
		},
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, OPTIONS",
	}))

	app.Get("/filter", filterHandler)

	log.Fatal(app.Listen(":9001"))
}

func filterHandler(c *fiber.Ctx) error {
	queryString := string(c.Request().URI().QueryString())
	targetURL := rustSource
	if queryString != "" {
		targetURL = fmt.Sprintf("%s?%s", rustSource, queryString)
	}

	c.Request().Header.Set("Host", targetHost)
	c.Request().Header.Set("X-Forwarded-For", c.IP())
	c.Request().Header.Del("Referer")

	return proxy.Do(c, targetURL)
}