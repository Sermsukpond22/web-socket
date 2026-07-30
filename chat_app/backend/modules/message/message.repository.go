package message

import (
	"chat_app/backend/models"

	"gorm.io/gorm"
)

type MessageRepository interface {
	SaveMessage(msg *models.Message) error
	GetChatHistory(userAID, userBID uint) ([]models.Message, error)
	GetChatHistoryPaginated(userAID, userBID uint, limit int, beforeID uint) ([]models.Message, error)
	GetLatestMessageID(userAID, userBID uint) (uint, error)
	GetUnreadCounts(userID uint) (map[uint]int64, error)
	MarkMessagesAsRead(senderID, receiverID uint) error
	GetMessageByID(msgID uint) (*models.Message, error)
	UpdateMessage(msg *models.Message) error
	DeleteMessage(msgID uint) error
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
		Order("id ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) GetChatHistoryPaginated(userAID, userBID uint, limit int, beforeID uint) ([]models.Message, error) {
	var messages []models.Message
	query := r.db.Preload("Sender").Preload("Receiver").
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userAID, userBID, userBID, userAID)
	
	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}

	// Fetch DESC to get latest messages, then we reverse them
	err := query.Order("id DESC").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse array to maintain ASC order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *messageRepository) MarkMessagesAsRead(senderID, receiverID uint) error {
	return r.db.Model(&models.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = ?", senderID, receiverID, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": gorm.Expr("CURRENT_TIMESTAMP")}).Error
}

func (r *messageRepository) GetLatestMessageID(userAID, userBID uint) (uint, error) {
	var maxID uint
	err := r.db.Model(&models.Message{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userAID, userBID, userBID, userAID).
		Select("COALESCE(MAX(id), 0)").Scan(&maxID).Error
	return maxID, err
}

func (r *messageRepository) GetUnreadCounts(userID uint) (map[uint]int64, error) {
	// Count messages where receiver is userID and is_read is false
	// Group by sender_id
	type Result struct {
		SenderID uint
		Count    int64
	}
	var results []Result
	err := r.db.Model(&models.Message{}).
		Select("sender_id, count(*) as count").
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Group("sender_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[uint]int64)
	for _, r := range results {
		counts[r.SenderID] = r.Count
	}
	return counts, nil
}

func (r *messageRepository) GetMessageByID(msgID uint) (*models.Message, error) {
	var msg models.Message
	err := r.db.First(&msg, msgID).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepository) UpdateMessage(msg *models.Message) error {
	return r.db.Save(msg).Error
}

func (r *messageRepository) DeleteMessage(msgID uint) error {
	return r.db.Model(&models.Message{}).Where("id = ?", msgID).Update("is_deleted", true).Error
}
