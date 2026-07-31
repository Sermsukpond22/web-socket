package models

import "time"

type RoomMember struct {
	RoomID   uint      `gorm:"primaryKey;autoIncrement:false" json:"room_id"`
	UserID   uint      `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	Role     string    `gorm:"type:varchar(50);default:'member'" json:"role"` // 'admin', 'member'
	Status   string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // 'pending', 'joined'
	JoinedAt time.Time `json:"joined_at"`

	Room Room `gorm:"foreignKey:RoomID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"room,omitempty"`
	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}
