package models

import (
	"time"
)

type FriendRequest struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FromUserID uint      `gorm:"not null;index:idx_from_to,unique;index:idx_from_user" json:"from_user_id"`
	ToUserID   uint      `gorm:"not null;index:idx_from_to,unique;index:idx_to_user" json:"to_user_id"`
	Status     string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // "pending", "accepted", "rejected"
	FromUser   User      `gorm:"foreignKey:FromUserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"from_user,omitempty"`
	ToUser     User      `gorm:"foreignKey:ToUserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"to_user,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Friendship struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index:idx_user_friend,unique;index:idx_user" json:"user_id"`
	FriendID  uint      `gorm:"not null;index:idx_user_friend,unique;index:idx_friend" json:"friend_id"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Friend    User      `gorm:"foreignKey:FriendID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"friend,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
