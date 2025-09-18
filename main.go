package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/gofiber/swagger" // swagger handler
	"github.com/gofiber/fiber/v2/middleware/logger"

	"workshop-be/internal/auth"
	"workshop-be/internal/db"
	"workshop-be/internal/middleware"
)

// @title Workshop BE API
// @version 1.0
// @description API for authentication workshop
// @BasePath /
// @schemes http

func main() {
	app := fiber.New()
	app.Use(logger.New())

	// Health
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	// Root
	app.Get("/", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message": "hello world"}) })

	// Env & DB init
	dbPath := os.Getenv("DB_PATH")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" { log.Fatal("JWT_SECRET not set") }
	// init database
	db.Init(dbPath, &auth.User{})

	// Auth routes
	authSvc := auth.NewService()
	authGroup := app.Group("/api/v1/auth")
	auth.RegisterRoutes(authGroup, authSvc)

	// Protected route
	authGroup.Get("/me", middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"email": c.Locals("user_email"),
			"id":    c.Locals("user_sub"),
		})
	})

	// Swagger endpoint (auto generated docs expected in internal/docs)
	// app.Get("/swagger/*", swagger.HandlerDefault) // enable after generating docs

	port := os.Getenv("PORT")
	if port == "" { port = "3000" }

	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Printf("shutting down server: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Gracefully shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Shutdown(); err != nil { log.Printf("error on shutdown: %v", err) }
	if err := db.Close(); err != nil { log.Printf("error closing db: %v", err) }
	_ = ctx
	log.Println("Server stopped")
}
