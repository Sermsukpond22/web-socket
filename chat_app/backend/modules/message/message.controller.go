package message

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type MessageController struct {
	messageService MessageService
}

func NewMessageController(messageService MessageService) *MessageController {
	return &MessageController{
		messageService: messageService,
	}
}

func (c *MessageController) GetChatHistory(ctx *fiber.Ctx) error {
	userIDVal := ctx.Locals("user_id")
	var userID uint
	switch v := userIDVal.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	case int:
		userID = uint(v)
	}

	if userID == 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized access",
		})
	}

	friendIDParam := ctx.Params("friend_id")
	friendID, err := strconv.ParseUint(friendIDParam, 10, 64)
	if err != nil || friendID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid friend ID",
		})
	}

	messages, err := c.messageService.GetChatHistory(userID, uint(friendID))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(messages)
}
