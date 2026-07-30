package models

import (
	"time"
)

type User struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username    string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Email       string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password    string     `gorm:"type:varchar(255);not null" json:"-"`
	DisplayName string     `gorm:"type:varchar(100)" json:"display_name"`
	Bio         string     `gorm:"type:text" json:"bio"`
	AvatarURL   string     `gorm:"type:varchar(255)" json:"avatar_url"`
	LastSeen    *time.Time `json:"last_seen"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
