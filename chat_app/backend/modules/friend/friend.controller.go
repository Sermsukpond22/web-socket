package friend

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"chat_app/backend/models"
	wsPkg "chat_app/backend/websocket"

	"github.com/gofiber/fiber/v2"
)

type FriendController struct {
	friendService FriendService
	wsHub         *wsPkg.Hub
}

func NewFriendController(friendService FriendService, wsHub *wsPkg.Hub) *FriendController {
	return &FriendController{
		friendService: friendService,
		wsHub:         wsHub,
	}
}

type SendRequestPayload struct {
	ToUserID   interface{} `json:"to_user_id"`
	ToUsername string      `json:"to_username"`
	ToUser     string      `json:"to_user"`
}

type AcceptRequestPayload struct {
	RequestID interface{} `json:"request_id"`
	ID        interface{} `json:"id"`
}

type PendingRequestDTO struct {
	RequestID  uint        `json:"request_id"`
	ID         uint        `json:"id"`
	FromUserID uint        `json:"from_user_id"`
	ToUserID   uint        `json:"to_user_id"`
	Status     string      `json:"status"`
	FromUser   models.User `json:"from_user"`
	CreatedAt  time.Time   `json:"created_at"`
}

func (c *FriendController) SendRequest(ctx *fiber.Ctx) error {
	userIDVal := ctx.Locals("user_id")
	if userIDVal == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid context user_id",
		})
	}

	var payload SendRequestPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	var input SendRequestInput
	var targetStr string

	// Extract to_user_id if present
	if payload.ToUserID != nil {
		switch v := payload.ToUserID.(type) {
		case float64:
			input.ToUserID = uint(v)
		case uint:
			input.ToUserID = v
		case int:
			input.ToUserID = uint(v)
		case string:
			if parsed, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64); err == nil {
				input.ToUserID = uint(parsed)
			} else {
				targetStr = strings.TrimSpace(v)
			}
		}
	}

	if payload.ToUsername != "" {
		input.ToUsername = strings.TrimSpace(payload.ToUsername)
	}

	if input.ToUserID == 0 && input.ToUsername == "" && payload.ToUser != "" {
		trimmedToUser := strings.TrimSpace(payload.ToUser)
		if parsed, err := strconv.ParseUint(trimmedToUser, 10, 64); err == nil {
			input.ToUserID = uint(parsed)
		} else {
			input.ToUsername = trimmedToUser
		}
	}

	var req *models.FriendRequest
	var err error

	if input.ToUserID > 0 || input.ToUsername != "" {
		req, err = c.friendService.SendFriendRequestByInput(userID, input)
	} else if targetStr != "" {
		req, err = c.friendService.SendFriendRequest(userID, targetStr)
	} else {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Must specify target user via to_user_id or to_username",
		})
	}

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if c.wsHub != nil && req != nil {
		senderUsername, _ := ctx.Locals("username").(string)
		c.wsHub.SendToUser(req.ToUserID, fiber.Map{
			"type":            "new_friend_request",
			"sender_id":       userID,
			"sender_username": senderUsername,
			"recipient_id":    req.ToUserID,
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Friend request sent",
		"request_id": req.ID,
		"status":     req.Status,
		"request":    req,
	})
}

func (c *FriendController) GetPendingRequests(ctx *fiber.Ctx) error {
	userIDVal := ctx.Locals("user_id")
	if userIDVal == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid context user_id",
		})
	}

	requests, err := c.friendService.GetPendingRequests(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	dtos := make([]PendingRequestDTO, 0, len(requests))
	for _, req := range requests {
		dtos = append(dtos, PendingRequestDTO{
			RequestID:  req.ID,
			ID:         req.ID,
			FromUserID: req.FromUserID,
			ToUserID:   req.ToUserID,
			Status:     req.Status,
			FromUser:   req.FromUser,
			CreatedAt:  req.CreatedAt,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(dtos)
}

func (c *FriendController) AcceptRequest(ctx *fiber.Ctx) error {
	userIDVal := ctx.Locals("user_id")
	if userIDVal == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid context user_id",
		})
	}

	var payload AcceptRequestPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	var requestID uint
	reqIDVal := payload.RequestID
	if reqIDVal == nil {
		reqIDVal = payload.ID
	}

	if reqIDVal != nil {
		switch v := reqIDVal.(type) {
		case float64:
			requestID = uint(v)
		case uint:
			requestID = v
		case int:
			requestID = uint(v)
		case string:
			if parsed, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64); err == nil {
				requestID = uint(parsed)
			}
		}
	}

	if requestID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid or missing request_id",
		})
	}

	req, err := c.friendService.AcceptFriendRequest(requestID, userID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if c.wsHub != nil && req != nil {
		senderUsername, _ := ctx.Locals("username").(string)
		c.wsHub.SendToUser(req.FromUserID, fiber.Map{
			"type":            "friend_request_accepted",
			"sender_id":       userID,
			"sender_username": senderUsername,
			"recipient_id":    req.FromUserID,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Friend request accepted",
		"status":     req.Status,
		"request_id": req.ID,
	})
}

func (c *FriendController) SearchUsers(ctx *fiber.Ctx) error {
	userIDVal := ctx.Locals("user_id")
	if userIDVal == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID, err := parseUintVal(userIDVal)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid context user_id",
		})
	}

	query := ctx.Query("q")
	users, err := c.friendService.SearchUsers(query, userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(users)
}

func (c *FriendController) GetFriends(ctx *fiber.Ctx) error {
	userIDVal := ctx.Locals("user_id")
	if userIDVal == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid context user_id",
		})
	}

	friends, err := c.friendService.GetFriends(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(friends)
}

func parseUintVal(v interface{}) (uint, error) {
	switch val := v.(type) {
	case float64:
		return uint(val), nil
	case uint:
		return val, nil
	case int:
		return uint(val), nil
	case string:
		p, err := strconv.ParseUint(strings.TrimSpace(val), 10, 64)
		if err != nil {
			return 0, err
		}
		return uint(p), nil
	default:
		return 0, fmt.Errorf("invalid type")
	}
}
