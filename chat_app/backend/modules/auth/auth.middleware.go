package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWTMiddleware(authService AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header missing",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format. Expected Bearer <token>",
			})
		}

		tokenStr := strings.TrimSpace(parts[1])
		_, claims, err := authService.ValidateToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		userIDVal, ok := claims["user_id"]
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		var userID uint
		switch v := userIDVal.(type) {
		case float64:
			userID = uint(v)
		case uint:
			userID = v
		case int:
			userID = uint(v)
		default:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid user_id type in claims",
			})
		}

		c.Locals("user_id", userID)
		if username, ok := claims["username"].(string); ok {
			c.Locals("username", username)
		}
		if email, ok := claims["email"].(string); ok {
			c.Locals("email", email)
		}

		return c.Next()
	}
}
