package room

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoomRoutes(app *fiber.App, controller *RoomController, jwtMiddleware fiber.Handler) {
	api := app.Group("/api/rooms", jwtMiddleware)
	
	api.Post("/", controller.CreateRoom)
	api.Get("/", controller.GetJoinedRooms)
	api.Post("/:id/invite", controller.InviteUser)
	api.Get("/invites", controller.GetPendingInvites)
	api.Post("/invites/:room_id/accept", controller.AcceptInvite)
}
