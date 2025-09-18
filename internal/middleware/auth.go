package middleware

import (
	"strings"

	"workshop-be/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{"code": "UNAUTHORIZED", "message": "missing or invalid token"}})
		}
		token := strings.TrimPrefix(h, "Bearer ")
		claims, err := auth.ParseToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{"code": "UNAUTHORIZED", "message": "invalid token"}})
		}
		c.Locals("user_email", claims.Email)
		c.Locals("user_sub", claims.Subject)
		return c.Next()
	}
}
