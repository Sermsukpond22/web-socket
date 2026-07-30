package auth

import (
	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(app *fiber.App, authController *AuthController, authService AuthService) {
	api := app.Group("/api")
	auth := api.Group("/auth")

	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Get("/me", JWTMiddleware(authService), authController.Me)
}
