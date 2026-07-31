package models

import (
	"time"
)

type Message struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID   uint       `gorm:"not null;index:idx_sender_receiver;index:idx_sender" json:"sender_id"`
	ReceiverID *uint      `gorm:"index:idx_sender_receiver;index:idx_receiver" json:"receiver_id,omitempty"`
	RoomID     *uint      `gorm:"index:idx_room" json:"room_id,omitempty"`
	ReplyToID  *uint      `json:"reply_to_id,omitempty"`
	Content    string     `gorm:"type:text;not null" json:"content"`
	Type       string     `gorm:"type:varchar(20);default:'text'" json:"type"` // "text", "image", "file"
	FileURL    string     `gorm:"type:varchar(500)" json:"file_url,omitempty"`
	IsDeleted  bool       `gorm:"default:false" json:"is_deleted"`
	IsEdited   bool       `gorm:"default:false" json:"is_edited"`
	IsRead     bool       `gorm:"default:false;index:idx_is_read" json:"is_read"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	CreatedAt  time.Time  `gorm:"index:idx_created_at" json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	Sender   User     `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sender,omitempty"`
	Receiver *User    `gorm:"foreignKey:ReceiverID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"receiver,omitempty"`
	Room     *Room    `gorm:"foreignKey:RoomID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"room,omitempty"`
	ReplyTo  *Message `gorm:"foreignKey:ReplyToID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"reply_to,omitempty"`
}
