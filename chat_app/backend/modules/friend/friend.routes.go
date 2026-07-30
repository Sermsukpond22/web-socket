package friend

import (
	"chat_app/backend/modules/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupFriendRoutes(app *fiber.App, friendController *FriendController, authService auth.AuthService) {
	api := app.Group("/api")
	api.Use(auth.JWTMiddleware(authService))
	friends := api.Group("/friends")

	friends.Get("/search", friendController.SearchUsers)
	friends.Post("/request", friendController.SendRequest)
	friends.Get("/pending", friendController.GetPendingRequests)
	friends.Post("/accept", friendController.AcceptRequest)
	friends.Delete("/reject", friendController.RejectRequest)
	friends.Get("/", friendController.GetFriends)
	friends.Delete("/:id", friendController.RemoveFriend)
}
