package models

import (
	"time"
)

type Message struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID   uint      `gorm:"not null;index:idx_sender_receiver;index:idx_sender" json:"sender_id"`
	ReceiverID uint      `gorm:"not null;index:idx_sender_receiver;index:idx_receiver" json:"receiver_id"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	CreatedAt  time.Time `gorm:"index:idx_created_at" json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Sender   User `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sender,omitempty"`
	Receiver User `gorm:"foreignKey:ReceiverID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"receiver,omitempty"`
}
