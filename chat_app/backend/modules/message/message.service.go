package message

import (
	"errors"
	"strings"

	"chat_app/backend/models"
	"chat_app/backend/modules/friend"
)

type MessageService interface {
	SendMessage(senderID, receiverID uint, content string) (*models.Message, error)
	GetChatHistory(userAID, userBID uint) ([]models.Message, error)
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

func (s *messageService) SendMessage(senderID, receiverID uint, content string) (*models.Message, error) {
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return nil, errors.New("message content cannot be empty")
	}

	if senderID == receiverID {
		return nil, errors.New("cannot send message to self")
	}

	areFriends, err := s.friendRepo.AreFriends(senderID, receiverID)
	if err != nil {
		return nil, err
	}
	if !areFriends {
		return nil, errors.New("users are not accepted friends")
	}

	msg := &models.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    trimmedContent,
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
