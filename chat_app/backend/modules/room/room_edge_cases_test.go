package room_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"chat_app/backend/models"
	"chat_app/backend/modules/message"
	"chat_app/backend/modules/room"
	"chat_app/backend/modules/user"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type testHarness struct {
	db          *gorm.DB
	app         *fiber.App
	roomService room.RoomService
	userRepo    user.UserRepository
	messageRepo message.MessageRepository
	admin       *models.User
	member      *models.User
	invitee     *models.User
	outsider    *models.User
	testRoom    *models.Room
}

func setupHTTPHarness(t *testing.T) *testHarness {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Room{}, &models.RoomMember{}, &models.Message{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	roomRepo := room.NewRoomRepository(db)
	userRepo := user.NewUserRepository(db)
	messageRepo := message.NewMessageRepository(db)
	roomService := room.NewRoomService(roomRepo, userRepo, messageRepo, nil)
	controller := room.NewRoomController(roomService, nil)

	app := fiber.New()

	// Dummy JWT middleware for testing that injects user_id from header X-Test-User-ID
	mockJWTMiddleware := func(c *fiber.Ctx) error {
		uidStr := c.Get("X-Test-User-ID")
		if uidStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing test user id"})
		}
		uid, err := strconv.ParseUint(uidStr, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid test user id"})
		}
		c.Locals("user_id", uint(uid))
		return c.Next()
	}

	room.SetupRoomRoutes(app, controller, mockJWTMiddleware)

	// Create users
	admin := &models.User{Username: "admin_user", Email: "admin@test.com", Password: "hash"}
	member := &models.User{Username: "member_user", Email: "member@test.com", Password: "hash"}
	invitee := &models.User{Username: "invitee_user", Email: "invitee@test.com", Password: "hash"}
	outsider := &models.User{Username: "outsider_user", Email: "outsider@test.com", Password: "hash"}

	userRepo.Create(admin)
	userRepo.Create(member)
	userRepo.Create(invitee)
	userRepo.Create(outsider)

	rm, err := roomService.CreateRoom("Harness Room", "", admin.ID)
	if err != nil {
		t.Fatalf("failed to create harness room: %v", err)
	}

	// Member invites & accepts
	roomService.InviteUser(rm.ID, admin.ID, member.ID)
	roomService.AcceptInvite(rm.ID, member.ID)

	// Invitee invited but NOT accepted (pending)
	roomService.InviteUser(rm.ID, admin.ID, invitee.ID)

	return &testHarness{
		db:          db,
		app:         app,
		roomService: roomService,
		userRepo:    userRepo,
		messageRepo: messageRepo,
		admin:       admin,
		member:      member,
		invitee:     invitee,
		outsider:    outsider,
		testRoom:    rm,
	}
}

func TestGetRoomMembers_HTTP_EdgeCases(t *testing.T) {
	h := setupHTTPHarness(t)
	roomIDStr := strconv.FormatUint(uint64(h.testRoom.ID), 10)

	// 1. Member requests members -> 200 OK
	req := httptest.NewRequest("GET", "/api/rooms/"+roomIDStr+"/members", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.member.ID), 10))
	resp, err := h.app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for member, got %v, err: %v", resp.StatusCode, err)
	}
	var members []models.RoomMember
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &members)
	if len(members) != 2 {
		t.Errorf("expected 2 joined members, got %d", len(members))
	}

	// 2. Outsider requests members -> 400 Bad Request
	req = httptest.NewRequest("GET", "/api/rooms/"+roomIDStr+"/members", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.outsider.ID), 10))
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for outsider, got %d", resp.StatusCode)
	}

	// 3. Pending invitee requests members -> 400 Bad Request
	req = httptest.NewRequest("GET", "/api/rooms/"+roomIDStr+"/members", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.invitee.ID), 10))
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for pending invitee, got %d", resp.StatusCode)
	}

	// 4. Invalid room ID parameter -> 400 Bad Request
	req = httptest.NewRequest("GET", "/api/rooms/invalid_id/members", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.member.ID), 10))
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid room ID, got %d", resp.StatusCode)
	}

	// 5. Non-existent room ID -> 400 Bad Request
	req = httptest.NewRequest("GET", "/api/rooms/999999/members", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.member.ID), 10))
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for non-existent room ID, got %d", resp.StatusCode)
	}
}

func TestLeaveRoom_HTTP_EdgeCases(t *testing.T) {
	h := setupHTTPHarness(t)
	roomIDStr := strconv.FormatUint(uint64(h.testRoom.ID), 10)

	// 1. Pending invitee tries to leave -> 400 Bad Request
	req := httptest.NewRequest("POST", "/api/rooms/"+roomIDStr+"/leave", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.invitee.ID), 10))
	resp, _ := h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for pending invitee leaving, got %d", resp.StatusCode)
	}

	// 2. Outsider tries to leave -> 400 Bad Request
	req = httptest.NewRequest("POST", "/api/rooms/"+roomIDStr+"/leave", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.outsider.ID), 10))
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for outsider leaving, got %d", resp.StatusCode)
	}

	// 3. Invalid room ID parameter -> 400 Bad Request
	req = httptest.NewRequest("POST", "/api/rooms/invalid_id/leave", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.member.ID), 10))
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid room ID, got %d", resp.StatusCode)
	}

	// 4. Member leaves -> 200 OK
	req = httptest.NewRequest("POST", "/api/rooms/"+roomIDStr+"/leave", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.member.ID), 10))
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for member leaving, got %d", resp.StatusCode)
	}

	// Verify system message created
	var sysMsgs []models.Message
	h.db.Where("room_id = ? AND type = ?", h.testRoom.ID, "system").Find(&sysMsgs)
	if len(sysMsgs) == 0 {
		t.Errorf("expected system message created after member left")
	} else if sysMsgs[0].Content != "member_user ออกจากกลุ่ม" {
		t.Errorf("unexpected system message content: %s", sysMsgs[0].Content)
	}

	// 5. Member tries to leave again (now non-member) -> 400 Bad Request
	req = httptest.NewRequest("POST", "/api/rooms/"+roomIDStr+"/leave", nil)
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(h.member.ID), 10))
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for former member leaving again, got %d", resp.StatusCode)
	}
}

func TestRemoveMember_HTTP_EdgeCases(t *testing.T) {
	h := setupHTTPHarness(t)
	roomIDStr := strconv.FormatUint(uint64(h.testRoom.ID), 10)
	memberIDStr := strconv.FormatUint(uint64(h.member.ID), 10)
	adminIDStr := strconv.FormatUint(uint64(h.admin.ID), 10)
	inviteeIDStr := strconv.FormatUint(uint64(h.invitee.ID), 10)
	outsiderIDStr := strconv.FormatUint(uint64(h.outsider.ID), 10)

	// 1. Admin attempts to remove self -> 400 Bad Request
	req := httptest.NewRequest("DELETE", "/api/rooms/"+roomIDStr+"/members/"+adminIDStr, nil)
	req.Header.Set("X-Test-User-ID", adminIDStr)
	resp, _ := h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for admin removing self, got %d", resp.StatusCode)
	}

	// 2. Non-admin member attempts to remove admin -> 400 Bad Request
	req = httptest.NewRequest("DELETE", "/api/rooms/"+roomIDStr+"/members/"+adminIDStr, nil)
	req.Header.Set("X-Test-User-ID", memberIDStr)
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for non-admin removing member, got %d", resp.StatusCode)
	}

	// 3. Outsider attempts to remove member -> 400 Bad Request
	req = httptest.NewRequest("DELETE", "/api/rooms/"+roomIDStr+"/members/"+memberIDStr, nil)
	req.Header.Set("X-Test-User-ID", outsiderIDStr)
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for outsider removing member, got %d", resp.StatusCode)
	}

	// 4. Admin attempts to remove outsider -> 400 Bad Request
	req = httptest.NewRequest("DELETE", "/api/rooms/"+roomIDStr+"/members/"+outsiderIDStr, nil)
	req.Header.Set("X-Test-User-ID", adminIDStr)
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for admin removing outsider, got %d", resp.StatusCode)
	}

	// 5. Admin attempts to remove pending invitee -> 400 Bad Request
	req = httptest.NewRequest("DELETE", "/api/rooms/"+roomIDStr+"/members/"+inviteeIDStr, nil)
	req.Header.Set("X-Test-User-ID", adminIDStr)
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for admin removing pending invitee, got %d", resp.StatusCode)
	}

	// 6. Invalid room ID parameter -> 400 Bad Request
	req = httptest.NewRequest("DELETE", "/api/rooms/invalid_id/members/"+memberIDStr, nil)
	req.Header.Set("X-Test-User-ID", adminIDStr)
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid room ID parameter, got %d", resp.StatusCode)
	}

	// 7. Invalid user ID parameter -> 400 Bad Request
	req = httptest.NewRequest("DELETE", "/api/rooms/"+roomIDStr+"/members/invalid_id", nil)
	req.Header.Set("X-Test-User-ID", adminIDStr)
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid user ID parameter, got %d", resp.StatusCode)
	}

	// 8. Admin removes member -> 200 OK
	req = httptest.NewRequest("DELETE", "/api/rooms/"+roomIDStr+"/members/"+memberIDStr, nil)
	req.Header.Set("X-Test-User-ID", adminIDStr)
	resp, _ = h.app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for admin removing member, got %d", resp.StatusCode)
	}

	// Verify system message created
	var sysMsgs []models.Message
	h.db.Where("room_id = ? AND type = ?", h.testRoom.ID, "system").Find(&sysMsgs)
	if len(sysMsgs) == 0 {
		t.Errorf("expected system message created after member removed")
	} else if sysMsgs[len(sysMsgs)-1].Content != "member_user ถูกลบออกจากกลุ่ม" {
		t.Errorf("unexpected system message content: %s", sysMsgs[len(sysMsgs)-1].Content)
	}
}
