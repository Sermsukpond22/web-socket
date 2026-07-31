package room

import (
	"chat_app/backend/models"
	"gorm.io/gorm"
)

type RoomRepository interface {
	CreateRoom(room *models.Room) error
	CreateRoomMember(member *models.RoomMember) error
	GetRoomsByUserID(userID uint) ([]models.Room, error)
	GetRoomByID(id uint) (*models.Room, error)
	GetPendingInvites(userID uint) ([]models.RoomMember, error)
	UpdateRoomMember(member *models.RoomMember) error
	GetRoomMember(roomID, userID uint) (*models.RoomMember, error)
	GetJoinedRoomMembers(roomID uint) ([]models.RoomMember, error)
	GetRoomMembersWithUser(roomID uint) ([]models.RoomMember, error)
	RemoveRoomMember(roomID, userID uint) error
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db}
}

func (r *roomRepository) CreateRoom(room *models.Room) error {
	return r.db.Create(room).Error
}

func (r *roomRepository) CreateRoomMember(member *models.RoomMember) error {
	return r.db.Create(member).Error
}

func (r *roomRepository) GetRoomsByUserID(userID uint) ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.Joins("JOIN room_members ON room_members.room_id = rooms.id").
		Where("room_members.user_id = ? AND room_members.status = ?", userID, "joined").
		Find(&rooms).Error
	return rooms, err
}

func (r *roomRepository) GetRoomByID(id uint) (*models.Room, error) {
	var room models.Room
	err := r.db.First(&room, id).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) GetPendingInvites(userID uint) ([]models.RoomMember, error) {
	var members []models.RoomMember
	err := r.db.Preload("Room").Where("user_id = ? AND status = ?", userID, "pending").Find(&members).Error
	return members, err
}

func (r *roomRepository) UpdateRoomMember(member *models.RoomMember) error {
	return r.db.Save(member).Error
}

func (r *roomRepository) GetRoomMember(roomID, userID uint) (*models.RoomMember, error) {
	var member models.RoomMember
	err := r.db.Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *roomRepository) GetJoinedRoomMembers(roomID uint) ([]models.RoomMember, error) {
	var members []models.RoomMember
	err := r.db.Where("room_id = ? AND status = ?", roomID, "joined").Find(&members).Error
	return members, err
}

func (r *roomRepository) GetRoomMembersWithUser(roomID uint) ([]models.RoomMember, error) {
	var members []models.RoomMember
	err := r.db.Preload("User").Where("room_id = ? AND status = ?", roomID, "joined").Find(&members).Error
	return members, err
}

func (r *roomRepository) RemoveRoomMember(roomID, userID uint) error {
	return r.db.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&models.RoomMember{}).Error
}
