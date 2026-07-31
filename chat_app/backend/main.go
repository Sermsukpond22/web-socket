package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat_app/backend/config"
	"chat_app/backend/models"
	"chat_app/backend/modules/auth"
	"chat_app/backend/modules/friend"
	"chat_app/backend/modules/message"
	"chat_app/backend/modules/room"
	"chat_app/backend/modules/user"
	wsPkg "chat_app/backend/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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
	if err := db.AutoMigrate(&models.User{}, &models.FriendRequest{}, &models.Friendship{}, &models.Message{}, &models.Room{}, &models.RoomMember{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// Patch messages table to allow NULL for receiver_id (for room messages)
	// Some older MySQL migrations might have set this to NOT NULL.
	if db.Dialector.Name() == "mysql" {
		db.Exec("ALTER TABLE messages MODIFY receiver_id bigint(20) unsigned NULL;")
	}

	log.Println("Database migration completed.")

	// JWT Secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	// Initialize Repositories, Services, and Controllers
	userRepo := user.NewUserRepository(db)
	friendRepo := friend.NewFriendRepository(db)
	messageRepo := message.NewMessageRepository(db)
	roomRepo := room.NewRoomRepository(db)

	authService := auth.NewAuthService(userRepo, jwtSecret)
	friendService := friend.NewFriendService(friendRepo, userRepo)
	messageService := message.NewMessageService(messageRepo, friendRepo)
	roomService := room.NewRoomService(roomRepo)

	wsHub := wsPkg.NewHub()
	wsHub.SetRoomRepo(roomRepo)
	wsHandler := wsPkg.NewWSHandler(wsHub, messageService, friendService, userRepo)

	authController := auth.NewAuthController(authService)
	friendController := friend.NewFriendController(friendService, wsHub)
	messageController := message.NewMessageController(messageService)
	userController := user.NewUserController(userRepo)
	roomController := room.NewRoomController(roomService, wsHub)

	// Create Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Real-Time Chat App Backend",
	})

	// Use Middlewares
	app.Use(logger.New())

	// Rate Limiter
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))

	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:5173, http://localhost:3000, http://127.0.0.1:5173, http://127.0.0.1:3000"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
	}))

	// Serve Static Files for Uploads
	os.MkdirAll("./uploads", os.ModePerm)
	app.Static("/uploads", "./uploads")

	// Setup Routes
	auth.SetupAuthRoutes(app, authController, authService)
	friend.SetupFriendRoutes(app, friendController, authService)
	message.SetupMessageRoutes(app, messageController, wsHandler, authService)
	user.RegisterUserRoutes(app, userController, auth.JWTMiddleware(authService))
	room.SetupRoomRoutes(app, roomController, auth.JWTMiddleware(authService))

	// Get Port from env or default to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s...", port)
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("Gracefully shutting down...")
	_ = app.ShutdownWithTimeout(10 * time.Second)
}
