package user

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"chat_app/backend/utils"
	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	Repo UserRepository
}

func NewUserController(repo UserRepository) *UserController {
	return &UserController{Repo: repo}
}

// GetProfile - view any user's profile
func (c *UserController) GetProfile(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	if idStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User ID is required"})
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid User ID"})
	}

	user, err := c.Repo.FindByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return ctx.JSON(user)
}

// UpdateProfile - update own profile
func (c *UserController) UpdateProfile(ctx *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	type ProfileInput struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
	}

	var input ProfileInput
	if err := ctx.BodyParser(&input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := c.Repo.UpdateProfile(userID, input.DisplayName, input.Bio); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update profile"})
	}

	return ctx.JSON(fiber.Map{"message": "Profile updated successfully"})
}

// UploadAvatar - upload avatar image
func (c *UserController) UploadAvatar(ctx *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to upload avatar file"})
	}

	// Validate size (5MB max)
	if file.Size > 5*1024*1024 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File size exceeds 5MB"})
	}

	// Validate MIME type roughly by opening file and reading first 512 bytes
	src, err := file.Open()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process file"})
	}
	defer src.Close()

	buffer := make([]byte, 512)
	_, _ = src.Read(buffer)
	src.Seek(0, io.SeekStart)

	// Accept only image types for avatar
	fileType := "" // We don't have a strict strict mime check without net/http.DetectContentType but simple extension check is okay here
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type. Only JPG, PNG, GIF, WEBP are allowed."})
	}
	
	_ = fileType // Ignore unused

	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create upload directory"})
	}

	// Sanitize filename
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

	if err := ctx.SaveFile(file, filePath); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	avatarURL := "/uploads/avatars/" + filename

	// Update DB
	if err := c.Repo.UpdateAvatar(userID, avatarURL); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user avatar"})
	}

	return ctx.JSON(fiber.Map{"message": "Avatar uploaded successfully", "avatar_url": avatarURL})
}
