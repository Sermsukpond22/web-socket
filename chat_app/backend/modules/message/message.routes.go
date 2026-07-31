package message

import (
	"chat_app/backend/modules/auth"
	wsPkg "chat_app/backend/websocket"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func SetupMessageRoutes(app *fiber.App, msgController *MessageController, wsHandler *wsPkg.WSHandler, authService auth.AuthService) {
	// REST API Routes
	api := app.Group("/api/messages")
	api.Use(auth.JWTMiddleware(authService))
	api.Get("/unread", msgController.GetUnreadCounts)
	api.Post("/upload", msgController.UploadFile)
	api.Get("/room/:room_id", msgController.GetRoomMessages)
	api.Get("/:friend_id", msgController.GetChatHistory)
	api.Post("/read/:friend_id", msgController.ReadMessages)

	// WebSocket Middleware for /ws
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			tokenStr := c.Query("token")
			if tokenStr == "" {
				authHeader := c.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
				} else {
					tokenStr = authHeader
				}
			}
			tokenStr = strings.TrimSpace(tokenStr)
			if tokenStr == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Missing token",
				})
			}

			_, claims, err := authService.ValidateToken(tokenStr)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid token",
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
			}

			c.Locals("user_id", userID)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket Route
	app.Get("/ws", websocket.New(wsHandler.HandleConnection))
}
