package user

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(router fiber.Router, userController *UserController, authMiddleware fiber.Handler) {
	userGroup := router.Group("/api/users")

	// Apply auth middleware to all user routes
	userGroup.Use(authMiddleware)

	userGroup.Get("/:id/profile", userController.GetProfile)
	userGroup.Put("/profile", userController.UpdateProfile)
	userGroup.Post("/avatar", userController.UploadAvatar)
}
