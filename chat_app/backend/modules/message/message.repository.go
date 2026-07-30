package message

import (
	"chat_app/backend/models"

	"gorm.io/gorm"
)

type MessageRepository interface {
	SaveMessage(msg *models.Message) error
	GetChatHistory(userAID, userBID uint) ([]models.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) SaveMessage(msg *models.Message) error {
	if err := r.db.Create(msg).Error; err != nil {
		return err
	}
	return r.db.Preload("Sender").Preload("Receiver").First(msg, msg.ID).Error
}

func (r *messageRepository) GetChatHistory(userAID, userBID uint) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("Sender").Preload("Receiver").
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userAID, userBID, userBID, userAID).
		Order("created_at ASC, id ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
