package message

import (
	"chat_app/backend/models"
	"chat_app/backend/utils"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
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

	limitStr := ctx.Query("limit", "50")
	beforeIDStr := ctx.Query("before_id", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	beforeID, err := strconv.ParseUint(beforeIDStr, 10, 64)
	if err != nil {
		beforeID = 0
	}

	var messages []models.Message
	if limit < 1000 { // If limit is provided, use paginated (default 50)
		messages, err = c.messageService.GetChatHistoryPaginated(userID, uint(friendID), limit, uint(beforeID))
	} else {
		messages, err = c.messageService.GetChatHistory(userID, uint(friendID))
	}

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(messages)
}

func (c *MessageController) ReadMessages(ctx *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
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

	err = c.messageService.MarkMessagesAsRead(uint(friendID), userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark messages as read",
		})
	}

	return ctx.JSON(fiber.Map{"success": true})
}

func (c *MessageController) GetUnreadCounts(ctx *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized access",
		})
	}

	counts, err := c.messageService.GetUnreadCounts(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get unread counts",
		})
	}

	return ctx.JSON(counts)
}

func (c *MessageController) UploadFile(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
	}

	// Max size 5MB
	if file.Size > 5*1024*1024 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File size exceeds 5MB limit"})
	}

	// Validate MIME type
	mimeType := file.Header.Get("Content-Type")
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" && mimeType != "application/pdf" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type"})
	}

	// Sanitize filename
	ext := filepath.Ext(file.Filename)
	safeFilename := strings.ReplaceAll(strings.ToLower(filepath.Base(file.Filename[:len(file.Filename)-len(ext)])), " ", "_")
	reg := regexp.MustCompile("[^a-zA-Z0-9_]+")
	safeFilename = reg.ReplaceAllString(safeFilename, "")
	if safeFilename == "" {
		safeFilename = "file"
	}

	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + safeFilename + ext
	os.MkdirAll("./uploads", os.ModePerm)
	err = ctx.SaveFile(file, "./uploads/"+filename)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	return ctx.JSON(fiber.Map{
		"file_url": "/uploads/" + filename,
	})
}
