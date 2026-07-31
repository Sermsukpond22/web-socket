package room

import (
	"errors"
	"fmt"

	"chat_app/backend/models"
	"chat_app/backend/modules/message"
	"chat_app/backend/modules/user"
	wsPkg "chat_app/backend/websocket"

	"github.com/gofiber/fiber/v2"
)

type RoomService interface {
	CreateRoom(name, avatarURL string, creatorID uint) (*models.Room, error)
	GetJoinedRooms(userID uint) ([]models.Room, error)
	InviteUser(roomID, inviterID, inviteeID uint) error
	GetPendingInvites(userID uint) ([]models.RoomMember, error)
	AcceptInvite(roomID, userID uint) error
	GetRoomMembers(roomID, userID uint) ([]models.RoomMember, error)
	LeaveRoom(roomID, userID uint) error
	RemoveMember(roomID, adminID, targetUserID uint) error
}

type roomService struct {
	repo        RoomRepository
	userRepo    user.UserRepository
	messageRepo message.MessageRepository
	wsHub       *wsPkg.Hub
}

func NewRoomService(repo RoomRepository, userRepo user.UserRepository, messageRepo message.MessageRepository, wsHub *wsPkg.Hub) RoomService {
	return &roomService{
		repo:        repo,
		userRepo:    userRepo,
		messageRepo: messageRepo,
		wsHub:       wsHub,
	}
}

func (s *roomService) CreateRoom(name, avatarURL string, creatorID uint) (*models.Room, error) {
	room := &models.Room{
		Name:      name,
		AvatarURL: avatarURL,
		CreatedBy: creatorID,
	}

	if err := s.repo.CreateRoom(room); err != nil {
		return nil, err
	}

	member := &models.RoomMember{
		RoomID: room.ID,
		UserID: creatorID,
		Role:   "admin",
		Status: "joined",
	}

	if err := s.repo.CreateRoomMember(member); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *roomService) GetJoinedRooms(userID uint) ([]models.Room, error) {
	return s.repo.GetRoomsByUserID(userID)
}

func (s *roomService) InviteUser(roomID, inviterID, inviteeID uint) error {
	inviterMember, err := s.repo.GetRoomMember(roomID, inviterID)
	if err != nil {
		return errors.New("inviter is not in the room")
	}

	if inviterMember.Role != "admin" {
		return errors.New("only admins can invite")
	}

	existingMember, err := s.repo.GetRoomMember(roomID, inviteeID)
	if err == nil && existingMember != nil {
		return errors.New("user is already invited or in the room")
	}

	member := &models.RoomMember{
		RoomID: roomID,
		UserID: inviteeID,
		Role:   "member",
		Status: "pending",
	}

	return s.repo.CreateRoomMember(member)
}

func (s *roomService) GetPendingInvites(userID uint) ([]models.RoomMember, error) {
	return s.repo.GetPendingInvites(userID)
}

func (s *roomService) AcceptInvite(roomID, userID uint) error {
	member, err := s.repo.GetRoomMember(roomID, userID)
	if err != nil {
		return errors.New("invite not found")
	}

	if member.Status != "pending" {
		return errors.New("invite already accepted or invalid")
	}

	member.Status = "joined"
	return s.repo.UpdateRoomMember(member)
}

func (s *roomService) GetRoomMembers(roomID, userID uint) ([]models.RoomMember, error) {
	member, err := s.repo.GetRoomMember(roomID, userID)
	if err != nil || member.Status != "joined" {
		return nil, errors.New("user is not a member of this room")
	}

	return s.repo.GetRoomMembersWithUser(roomID)
}

func (s *roomService) LeaveRoom(roomID, userID uint) error {
	member, err := s.repo.GetRoomMember(roomID, userID)
	if err != nil || member.Status != "joined" {
		return errors.New("user is not a member of this room")
	}

	var username string
	if s.userRepo != nil {
		userObj, err := s.userRepo.FindByID(userID)
		if err == nil && userObj != nil {
			username = userObj.Username
		}
	}

	if err := s.repo.RemoveRoomMember(roomID, userID); err != nil {
		return err
	}

	sysMsg := &models.Message{
		RoomID:   &roomID,
		SenderID: userID,
		Content:  fmt.Sprintf("%s ออกจากกลุ่ม", username),
		Type:     "system",
	}

	if s.messageRepo != nil {
		_ = s.messageRepo.SaveMessage(sysMsg)
	}

	if s.wsHub != nil {
		chatEvent := wsPkg.WSChatEvent{
			Type:      "chat",
			ID:        sysMsg.ID,
			SenderID:  sysMsg.SenderID,
			RoomID:    sysMsg.RoomID,
			Content:   sysMsg.Content,
			MsgType:   sysMsg.Type,
			CreatedAt: sysMsg.CreatedAt,
		}
		s.wsHub.SendToRoom(roomID, chatEvent)
		s.wsHub.SendToUser(userID, fiber.Map{"type": "room_left", "room_id": roomID})
	}

	return nil
}

func (s *roomService) RemoveMember(roomID, adminID, targetUserID uint) error {
	if adminID == targetUserID {
		return errors.New("admin cannot remove self")
	}

	adminMember, err := s.repo.GetRoomMember(roomID, adminID)
	if err != nil || adminMember.Status != "joined" {
		return errors.New("admin is not in the room")
	}

	if adminMember.Role != "admin" {
		return errors.New("only admins can remove members")
	}

	targetMember, err := s.repo.GetRoomMember(roomID, targetUserID)
	if err != nil || targetMember.Status != "joined" {
		return errors.New("target user is not in the room")
	}

	var targetUsername string
	if s.userRepo != nil {
		targetUser, err := s.userRepo.FindByID(targetUserID)
		if err == nil && targetUser != nil {
			targetUsername = targetUser.Username
		}
	}

	if err := s.repo.RemoveRoomMember(roomID, targetUserID); err != nil {
		return err
	}

	sysMsg := &models.Message{
		RoomID:   &roomID,
		SenderID: adminID,
		Content:  fmt.Sprintf("%s ถูกลบออกจากกลุ่ม", targetUsername),
		Type:     "system",
	}

	if s.messageRepo != nil {
		_ = s.messageRepo.SaveMessage(sysMsg)
	}

	if s.wsHub != nil {
		chatEvent := wsPkg.WSChatEvent{
			Type:      "chat",
			ID:        sysMsg.ID,
			SenderID:  sysMsg.SenderID,
			RoomID:    sysMsg.RoomID,
			Content:   sysMsg.Content,
			MsgType:   sysMsg.Type,
			CreatedAt: sysMsg.CreatedAt,
		}
		s.wsHub.SendToRoom(roomID, chatEvent)
		s.wsHub.SendToUser(targetUserID, fiber.Map{"type": "room_kicked", "room_id": roomID})
	}

	return nil
}
