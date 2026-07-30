package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetUserIDFromContext(c *fiber.Ctx) (uint, error) {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return 0, fmt.Errorf("user_id not found in context")
	}

	switch v := userIDVal.(type) {
	case uint:
		return v, nil
	case float64:
		return uint(v), nil
	default:
		return 0, fmt.Errorf("invalid user_id type in context")
	}
}
