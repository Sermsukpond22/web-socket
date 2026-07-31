package room

import (
	"strconv"
	"github.com/gofiber/fiber/v2"
	wsPkg "chat_app/backend/websocket"
)

type RoomController struct {
	service RoomService
	wsHub   *wsPkg.Hub
}

func NewRoomController(service RoomService, wsHub *wsPkg.Hub) *RoomController {
	return &RoomController{service, wsHub}
}

func (c *RoomController) CreateRoom(ctx *fiber.Ctx) error {
	var input struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	userID := ctx.Locals("user_id").(uint)

	room, err := c.service.CreateRoom(input.Name, input.AvatarURL, userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(room)
}

func (c *RoomController) GetJoinedRooms(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(uint)
	rooms, err := c.service.GetJoinedRooms(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(rooms)
}

func (c *RoomController) InviteUser(ctx *fiber.Ctx) error {
	roomIDParam := ctx.Params("id")
	roomID, err := strconv.ParseUint(roomIDParam, 10, 32)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid room ID"})
	}

	var input struct {
		UserID uint `json:"user_id"`
	}

	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	inviterID := ctx.Locals("user_id").(uint)

	if err := c.service.InviteUser(uint(roomID), inviterID, input.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if c.wsHub != nil {
		c.wsHub.SendToUser(input.UserID, fiber.Map{
			"type":    "room_invite",
			"room_id": uint(roomID),
		})
	}

	return ctx.JSON(fiber.Map{"message": "User invited"})
}

func (c *RoomController) GetPendingInvites(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(uint)
	invites, err := c.service.GetPendingInvites(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(invites)
}

func (c *RoomController) AcceptInvite(ctx *fiber.Ctx) error {
	roomIDParam := ctx.Params("room_id")
	roomID, err := strconv.ParseUint(roomIDParam, 10, 32)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid room ID"})
	}

	userID := ctx.Locals("user_id").(uint)

	if err := c.service.AcceptInvite(uint(roomID), userID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{"message": "Invite accepted"})
}
