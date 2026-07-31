package room

import (
	"chat_app/backend/models"
	"errors"
)

type RoomService interface {
	CreateRoom(name, avatarURL string, creatorID uint) (*models.Room, error)
	GetJoinedRooms(userID uint) ([]models.Room, error)
	InviteUser(roomID, inviterID, inviteeID uint) error
	GetPendingInvites(userID uint) ([]models.RoomMember, error)
	AcceptInvite(roomID, userID uint) error
}

type roomService struct {
	repo RoomRepository
}

func NewRoomService(repo RoomRepository) RoomService {
	return &roomService{repo}
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
