package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/leonfiasco/boomerang-go/database"
	"github.com/leonfiasco/boomerang-go/routes/auth"
)

func main() {
	database.LoadDB()

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	userGroup := app.Group("/user")

	auth.SetupRoutes(userGroup)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
