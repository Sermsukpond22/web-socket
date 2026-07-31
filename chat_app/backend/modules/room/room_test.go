package room_test

import (
	"testing"

	"chat_app/backend/models"
	"chat_app/backend/modules/message"
	"chat_app/backend/modules/room"
	"chat_app/backend/modules/user"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Room{}, &models.RoomMember{}, &models.Message{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestRoomMemberManagement(t *testing.T) {
	db := setupTestDB(t)

	roomRepo := room.NewRoomRepository(db)
	userRepo := user.NewUserRepository(db)
	messageRepo := message.NewMessageRepository(db)
	roomService := room.NewRoomService(roomRepo, userRepo, messageRepo, nil)

	// Create test users
	admin := &models.User{Username: "admin_user", Email: "admin@example.com", Password: "hashedpassword"}
	memberUser := &models.User{Username: "member_user", Email: "member@example.com", Password: "hashedpassword"}
	outsider := &models.User{Username: "outsider_user", Email: "outsider@example.com", Password: "hashedpassword"}

	if err := userRepo.Create(admin); err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}
	if err := userRepo.Create(memberUser); err != nil {
		t.Fatalf("failed to create memberUser: %v", err)
	}
	if err := userRepo.Create(outsider); err != nil {
		t.Fatalf("failed to create outsider: %v", err)
	}

	// Create room with admin as creator
	rm, err := roomService.CreateRoom("Test Room", "", admin.ID)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	// Invite memberUser and accept invite
	if err := roomService.InviteUser(rm.ID, admin.ID, memberUser.ID); err != nil {
		t.Fatalf("failed to invite memberUser: %v", err)
	}
	if err := roomService.AcceptInvite(rm.ID, memberUser.ID); err != nil {
		t.Fatalf("failed to accept invite for memberUser: %v", err)
	}

	// Test GetRoomMembers
	// 1. Outsider should fail
	_, err = roomService.GetRoomMembers(rm.ID, outsider.ID)
	if err == nil {
		t.Errorf("expected error when outsider requests room members, got nil")
	}

	// 2. Member should succeed and get 2 members with User preloaded
	members, err := roomService.GetRoomMembers(rm.ID, memberUser.ID)
	if err != nil {
		t.Fatalf("unexpected error getting room members: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("expected 2 room members, got %d", len(members))
	}
	foundAdminUser := false
	foundMemberUser := false
	for _, m := range members {
		if m.UserID == admin.ID && m.User.Username == "admin_user" {
			foundAdminUser = true
		}
		if m.UserID == memberUser.ID && m.User.Username == "member_user" {
			foundMemberUser = true
		}
	}
	if !foundAdminUser || !foundMemberUser {
		t.Errorf("expected preloaded users in room members, admin: %v, member: %v", foundAdminUser, foundMemberUser)
	}

	// Test RemoveMember
	// 1. Admin cannot remove self
	err = roomService.RemoveMember(rm.ID, admin.ID, admin.ID)
	if err == nil {
		t.Errorf("expected error when admin removes self, got nil")
	}

	// 2. Member cannot remove admin
	err = roomService.RemoveMember(rm.ID, memberUser.ID, admin.ID)
	if err == nil {
		t.Errorf("expected error when non-admin tries to remove member, got nil")
	}

	// 3. Admin removes memberUser
	err = roomService.RemoveMember(rm.ID, admin.ID, memberUser.ID)
	if err != nil {
		t.Fatalf("unexpected error removing member: %v", err)
	}

	// Verify memberUser is removed
	membersAfterRemove, err := roomService.GetRoomMembers(rm.ID, admin.ID)
	if err != nil {
		t.Fatalf("failed to get members after remove: %v", err)
	}
	if len(membersAfterRemove) != 1 {
		t.Errorf("expected 1 room member after remove, got %d", len(membersAfterRemove))
	}

	// Test LeaveRoom
	// memberUser is no longer in room, LeaveRoom should fail
	err = roomService.LeaveRoom(rm.ID, memberUser.ID)
	if err == nil {
		t.Errorf("expected error when non-member leaves room, got nil")
	}

	// Admin leaves room
	err = roomService.LeaveRoom(rm.ID, admin.ID)
	if err != nil {
		t.Fatalf("unexpected error when admin leaves room: %v", err)
	}

	// Verify room is empty
	membersAfterLeave, _ := roomRepo.GetRoomMembersWithUser(rm.ID)
	if len(membersAfterLeave) != 0 {
		t.Errorf("expected 0 members after admin left, got %d", len(membersAfterLeave))
	}
}
