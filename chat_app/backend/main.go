package main

import (
	"log"
	"os"

	"chat_app/backend/config"
	"chat_app/backend/models"
	"chat_app/backend/modules/auth"
	"chat_app/backend/modules/friend"
	"chat_app/backend/modules/message"
	"chat_app/backend/modules/user"
	wsPkg "chat_app/backend/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	// Initialize Database Connection
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Auto-migrate models
	log.Println("Running database migrations...")
	if err := db.AutoMigrate(&models.User{}, &models.FriendRequest{}, &models.Friendship{}, &models.Message{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	log.Println("Database migration completed.")

	// JWT Secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_jwt_secret_key_chat_app"
	}

	// Initialize Repositories, Services, and Controllers
	userRepo := user.NewUserRepository(db)
	friendRepo := friend.NewFriendRepository(db)
	messageRepo := message.NewMessageRepository(db)

	authService := auth.NewAuthService(userRepo, jwtSecret)
	friendService := friend.NewFriendService(friendRepo, userRepo)
	messageService := message.NewMessageService(messageRepo, friendRepo)

	wsHub := wsPkg.NewHub()
	wsHandler := wsPkg.NewWSHandler(wsHub, messageService, friendService)

	authController := auth.NewAuthController(authService)
	friendController := friend.NewFriendController(friendService, wsHub)
	messageController := message.NewMessageController(messageService)

	// Create Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Real-Time Chat App Backend",
	})

	// Use Middlewares
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, http://localhost:3000, http://127.0.0.1:5173, http://127.0.0.1:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
	}))

	// Setup Routes
	auth.SetupAuthRoutes(app, authController, authService)
	friend.SetupFriendRoutes(app, friendController, authService)
	message.SetupMessageRoutes(app, messageController, wsHandler, authService)

	// Get Port from env or default to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
