package auth

import (
	"chat_app/backend/utils"
	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	authService AuthService
}

func NewAuthController(authService AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (c *AuthController) Register(ctx *fiber.Ctx) error {
	var input RegisterInput
	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	user, err := c.authService.Register(input)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":    user,
		"message": "สมัครสมาชิกสำเร็จ",
	})
}

func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var input LoginInput
	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	user, token, refreshToken, err := c.authService.Login(input)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func (c *AuthController) Me(ctx *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	user, err := c.authService.GetUserByID(userID)
	if err != nil || user == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": user,
	})
}

func (c *AuthController) RefreshToken(ctx *fiber.Ctx) error {
	type RefreshInput struct {
		RefreshToken string `json:"refresh_token"`
	}
	var input RefreshInput
	if err := ctx.BodyParser(&input); err != nil || input.RefreshToken == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	_, claims, err := c.authService.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	userID := uint(claims["user_id"].(float64))
	user, err := c.authService.GetUserByID(userID)
	if err != nil || user == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	newToken, err := c.authService.GenerateToken(user)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": newToken,
	})
}
