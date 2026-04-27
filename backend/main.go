package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/WorkPlace/Postcube/backend/database"
	"github.com/WorkPlace/Postcube/backend/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	database.InitDB()

	app := fiber.New(fiber.Config{AppName: "Postcube API"})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     getEnv("ALLOWED_ORIGINS", "http://localhost:5116"),
		AllowMethods:     "GET,POST,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "postcube-backend"})
	})

	auth := app.Group("/api/auth")
	auth.Get("/login", handlers.Login)
	auth.Get("/callback", handlers.Callback)
	auth.Post("/logout", handlers.Logout)
	auth.Get("/me", handlers.AuthMiddleware, handlers.Me)

	public := app.Group("/api/public")
	public.Get("/box/:slug", handlers.GetPublicBox)
	public.Post(
		"/box/:slug/questions",
		limiter.New(limiter.Config{
			Max:        8,
			Expiration: time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP() + "|" + c.Params("slug")
			},
		}),
		handlers.SubmitAnonymousQuestion,
	)

	api := app.Group("/api", handlers.AuthMiddleware)
	api.Get("/box/me", handlers.GetMyBox)
	api.Patch("/box/me", handlers.UpdateMyBox)
	api.Get("/inbox", handlers.GetInbox)
	api.Patch("/inbox/questions/:id", handlers.UpdateInboxQuestion)
	api.Delete("/inbox/questions/:id", handlers.DeleteInboxQuestion)

	port := getEnv("PORT", "8113")
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	log.Printf("Postcube API listening on %s", port)
	log.Fatal(app.Listen(port))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
