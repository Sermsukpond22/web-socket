package message

import (
	"errors"
	"strings"

	"chat_app/backend/models"
	"chat_app/backend/modules/friend"
)

type MessageService interface {
	SendMessage(senderID uint, receiverID, roomID, replyToID *uint, content, msgType, fileURL string) (*models.Message, error)
	GetChatHistory(userAID, userBID uint) ([]models.Message, error)
	GetChatHistoryPaginated(userAID, userBID uint, limit int, beforeID uint) ([]models.Message, error)
	GetRoomChatHistory(roomID uint) ([]models.Message, error)
	GetRoomChatHistoryPaginated(roomID uint, limit int, beforeID uint) ([]models.Message, error)
	GetLatestMessageID(userAID, userBID uint) (uint, error)
	GetUnreadCounts(userID uint) (map[uint]int64, error)
	MarkMessagesAsRead(senderID, receiverID uint) error
	EditMessage(userID, msgID uint, newContent string) (*models.Message, error)
	DeleteMessage(userID, msgID uint) error
}

type messageService struct {
	messageRepo MessageRepository
	friendRepo  friend.FriendRepository
}

func NewMessageService(messageRepo MessageRepository, friendRepo friend.FriendRepository) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		friendRepo:  friendRepo,
	}
}

func (s *messageService) SendMessage(senderID uint, receiverID, roomID, replyToID *uint, content, msgType, fileURL string) (*models.Message, error) {
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" && fileURL == "" {
		return nil, errors.New("message content cannot be empty")
	}

	if msgType == "" {
		msgType = "text"
	}

	if receiverID != nil {
		if senderID == *receiverID {
			return nil, errors.New("cannot send message to self")
		}

		areFriends, err := s.friendRepo.AreFriends(senderID, *receiverID)
		if err != nil {
			return nil, err
		}
		if !areFriends {
			return nil, errors.New("users are not accepted friends")
		}
	} else if roomID == nil {
		return nil, errors.New("must specify either receiver_id or room_id")
	}
	msg := &models.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		RoomID:     roomID,
		ReplyToID:  replyToID,
		Content:    trimmedContent,
		Type:       msgType,
		FileURL:    fileURL,
	}

	if err := s.messageRepo.SaveMessage(msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *messageService) GetChatHistory(userAID, userBID uint) ([]models.Message, error) {
	areFriends, err := s.friendRepo.AreFriends(userAID, userBID)
	if err != nil {
		return nil, err
	}
	if !areFriends {
		return nil, errors.New("users are not accepted friends")
	}

	return s.messageRepo.GetChatHistory(userAID, userBID)
}

func (s *messageService) GetChatHistoryPaginated(userAID, userBID uint, limit int, beforeID uint) ([]models.Message, error) {
	areFriends, err := s.friendRepo.AreFriends(userAID, userBID)
	if err != nil {
		return nil, err
	}
	if !areFriends {
		return nil, errors.New("users are not accepted friends")
	}

	return s.messageRepo.GetChatHistoryPaginated(userAID, userBID, limit, beforeID)
}

func (s *messageService) GetRoomChatHistory(roomID uint) ([]models.Message, error) {
	return s.messageRepo.GetRoomChatHistory(roomID)
}

func (s *messageService) GetRoomChatHistoryPaginated(roomID uint, limit int, beforeID uint) ([]models.Message, error) {
	return s.messageRepo.GetRoomChatHistoryPaginated(roomID, limit, beforeID)
}

func (s *messageService) MarkMessagesAsRead(senderID, receiverID uint) error {
	err := s.messageRepo.MarkMessagesAsRead(senderID, receiverID)
	if err != nil {
		return err
	}

	maxID, err := s.messageRepo.GetLatestMessageID(senderID, receiverID)
	if err != nil {
		return err
	}

	if maxID > 0 {
		return s.friendRepo.UpdateLastReadMessageID(receiverID, senderID, maxID)
	}

	return nil
}

func (s *messageService) GetLatestMessageID(userAID, userBID uint) (uint, error) {
	return s.messageRepo.GetLatestMessageID(userAID, userBID)
}

func (s *messageService) GetUnreadCounts(userID uint) (map[uint]int64, error) {
	return s.messageRepo.GetUnreadCounts(userID)
}

func (s *messageService) EditMessage(userID, msgID uint, newContent string) (*models.Message, error) {
	msg, err := s.messageRepo.GetMessageByID(msgID)
	if err != nil {
		return nil, err
	}

	if msg.SenderID != userID {
		return nil, errors.New("unauthorized to edit this message")
	}

	trimmedContent := strings.TrimSpace(newContent)
	if trimmedContent == "" {
		return nil, errors.New("content cannot be empty")
	}

	msg.Content = trimmedContent
	msg.IsEdited = true

	if err := s.messageRepo.UpdateMessage(msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *messageService) DeleteMessage(userID, msgID uint) error {
	msg, err := s.messageRepo.GetMessageByID(msgID)
	if err != nil {
		return err
	}

	if msg.SenderID != userID {
		return errors.New("unauthorized to delete this message")
	}

	return s.messageRepo.DeleteMessage(msgID)
}
