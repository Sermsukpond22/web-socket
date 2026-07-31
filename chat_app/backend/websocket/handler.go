package websocket

import (
	"log"
	"time"

	"chat_app/backend/models"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type WSIncomingMessage struct {
	Type        string `json:"type"`
	ID          uint   `json:"id,omitempty"`
	ReceiverID  *uint  `json:"receiver_id,omitempty"`
	RoomID      *uint  `json:"room_id,omitempty"`
	ReplyToID   *uint  `json:"reply_to_id,omitempty"`
	RecipientID uint   `json:"recipient_id,omitempty"`
	Content     string `json:"content"`
	MsgType     string `json:"msg_type,omitempty"`
	FileURL     string `json:"file_url,omitempty"`
	IsTyping    *bool  `json:"is_typing,omitempty"`
}

type WSChatEvent struct {
	Type       string          `json:"type"`
	ID         uint            `json:"id"`
	SenderID   uint            `json:"sender_id"`
	ReceiverID *uint           `json:"receiver_id,omitempty"`
	RoomID     *uint           `json:"room_id,omitempty"`
	ReplyToID  *uint           `json:"reply_to_id,omitempty"`
	ReplyTo    *models.Message `json:"reply_to,omitempty"`
	Content    string          `json:"content"`
	MsgType    string          `json:"msg_type,omitempty"`
	FileURL    string          `json:"file_url,omitempty"`
	IsDeleted  bool            `json:"is_deleted,omitempty"`
	IsEdited   bool            `json:"is_edited,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type WSErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type MessageService interface {
	SendMessage(senderID uint, receiverID, roomID, replyToID *uint, content, msgType, fileURL string) (*models.Message, error)
	MarkMessagesAsRead(senderID, receiverID uint) error
	EditMessage(userID, msgID uint, content string) (*models.Message, error)
	DeleteMessage(userID, msgID uint) error
}

type FriendService interface {
	AreFriends(userAID, userBID uint) (bool, error)
	GetFriends(userID uint) ([]models.User, error)
}

type UserRepository interface {
	UpdateLastSeen(userID uint, lastSeen *time.Time) error
}

type WSHandler struct {
	hub            *Hub
	messageService MessageService
	friendService  FriendService
	userRepo       UserRepository
}

func NewWSHandler(hub *Hub, messageService MessageService, friendService FriendService, userRepo UserRepository) *WSHandler {
	return &WSHandler{
		hub:            hub,
		messageService: messageService,
		friendService:  friendService,
		userRepo:       userRepo,
	}
}

func (h *WSHandler) HandleConnection(c *websocket.Conn) {
	// fiber context isn't accessible directly in the websocket connection
	// Wait, we need to extract from c.Locals
	userIDVal := c.Locals("user_id")
	var userID uint
	switch v := userIDVal.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	case int:
		userID = uint(v)
	}

	if userID == 0 {
		_ = c.WriteJSON(WSErrorEvent{
			Type:    "error",
			Message: "Unauthorized websocket connection",
		})
		_ = c.Close()
		return
	}

	client := NewClient(userID, c, h.hub)
	h.hub.Register(client)

	// Update last_seen to null (online)
	h.userRepo.UpdateLastSeen(userID, nil)
	// Broadcast user online status
	h.broadcastUserStatus(userID, true, nil)

	// Send current statuses of friends to the newly connected user
	friends, err := h.friendService.GetFriends(userID)
	if err == nil {
		for _, f := range friends {
			isOnline := h.hub.IsUserOnline(f.ID)
			_ = client.WriteJSON(fiber.Map{
				"type":      "user_status",
				"user_id":   f.ID,
				"is_online": isOnline,
				"last_seen": f.LastSeen,
			})
		}
	}

	defer func() {
		h.hub.Unregister(client)
		// Update last_seen to now (offline)
		now := time.Now()
		h.userRepo.UpdateLastSeen(userID, &now)
		// Broadcast user offline status
		h.broadcastUserStatus(userID, false, &now)
	}()

	log.Printf("[WS] Client connected: user_id=%d", userID)

	for {
		var incoming WSIncomingMessage
		err := c.ReadJSON(&incoming)
		if err != nil {
			log.Printf("[WS] Client disconnected or read error: user_id=%d, err=%v", userID, err)
			break
		}

		h.handleMessage(client, incoming)
	}
}

func (h *WSHandler) handleMessage(client *Client, incoming WSIncomingMessage) {
	// Treat empty type as "chat"
	msgType := incoming.Type
	if msgType == "" {
		msgType = "chat"
	}

	switch msgType {
	case "ping":
		_ = client.WriteJSON(fiber.Map{"type": "pong"})
	case "typing":
		var receiverID uint
		if incoming.ReceiverID != nil {
			receiverID = *incoming.ReceiverID
		} else if incoming.RecipientID != 0 {
			receiverID = incoming.RecipientID
		}
		if receiverID == 0 || receiverID == client.UserID {
			return
		}

		if h.friendService != nil {
			areFriends, err := h.friendService.AreFriends(client.UserID, receiverID)
			if err != nil || !areFriends {
				return
			}
		}

		isTyping := true
		if incoming.IsTyping != nil {
			isTyping = *incoming.IsTyping
		}

		h.hub.SendToUser(receiverID, fiber.Map{
			"type":      "typing",
			"sender_id": client.UserID,
			"is_typing": isTyping,
		})
	case "chat":
		savedMsg, err := h.messageService.SendMessage(client.UserID, incoming.ReceiverID, incoming.RoomID, incoming.ReplyToID, incoming.Content, incoming.MsgType, incoming.FileURL)
		if err != nil {
			_ = client.WriteJSON(WSErrorEvent{
				Type:    "error",
				Message: err.Error(),
			})
			return
		}

		chatEvent := WSChatEvent{
			Type:       "chat",
			ID:         savedMsg.ID,
			SenderID:   savedMsg.SenderID,
			ReceiverID: savedMsg.ReceiverID,
			RoomID:     savedMsg.RoomID,
			ReplyToID:  savedMsg.ReplyToID,
			ReplyTo:    savedMsg.ReplyTo,
			Content:    savedMsg.Content,
			MsgType:    savedMsg.Type,
			FileURL:    savedMsg.FileURL,
			CreatedAt:  savedMsg.CreatedAt,
		}

		_ = client.WriteJSON(chatEvent)

		if savedMsg.RoomID != nil {
			h.hub.SendToRoom(*savedMsg.RoomID, chatEvent)
		} else if savedMsg.ReceiverID != nil && *savedMsg.ReceiverID != client.UserID {
			h.hub.SendToUser(*savedMsg.ReceiverID, chatEvent)
		}
	case "read":
		var senderID uint
		if incoming.ReceiverID != nil {
			senderID = *incoming.ReceiverID
		}
		if senderID == 0 || senderID == client.UserID {
			return
		}

		err := h.messageService.MarkMessagesAsRead(senderID, client.UserID)
		if err != nil {
			return
		}

		// Notify the original sender that their messages were read
		h.hub.SendToUser(senderID, fiber.Map{
			"type":      "read_receipt",
			"reader_id": client.UserID,
		})

	case "edit_message":
		msgID := incoming.ID
		editedMsg, err := h.messageService.EditMessage(client.UserID, msgID, incoming.Content)
		if err != nil {
			_ = client.WriteJSON(WSErrorEvent{Type: "error", Message: err.Error()})
			return
		}

		editEvent := WSChatEvent{
			Type:       "message_edited",
			ID:         editedMsg.ID,
			SenderID:   editedMsg.SenderID,
			ReceiverID: editedMsg.ReceiverID,
			RoomID:     editedMsg.RoomID,
			Content:    editedMsg.Content,
			MsgType:    editedMsg.Type,
			FileURL:    editedMsg.FileURL,
			IsEdited:   editedMsg.IsEdited,
			IsDeleted:  editedMsg.IsDeleted,
			CreatedAt:  editedMsg.CreatedAt,
		}
		_ = client.WriteJSON(editEvent)
		if editedMsg.RoomID != nil {
			h.hub.SendToRoom(*editedMsg.RoomID, editEvent)
		} else if editedMsg.ReceiverID != nil {
			h.hub.SendToUser(*editedMsg.ReceiverID, editEvent)
		}

	case "delete_message":
		msgID := incoming.ID
		err := h.messageService.DeleteMessage(client.UserID, msgID)
		if err != nil {
			_ = client.WriteJSON(WSErrorEvent{Type: "error", Message: err.Error()})
			return
		}

		deleteEvent := fiber.Map{
			"type": "message_deleted",
			"id":   msgID,
		}
		_ = client.WriteJSON(deleteEvent)
		// To find the receiver, we should ideally fetch the message before deleting.
		// Since we deleted it (soft delete), we can just broadcast to receiver if we had it.
		// For now we can require the client to pass receiver_id in incoming message.
		if incoming.RoomID != nil && *incoming.RoomID > 0 {
			h.hub.SendToRoom(*incoming.RoomID, deleteEvent)
		} else if incoming.ReceiverID != nil && *incoming.ReceiverID > 0 {
			h.hub.SendToUser(*incoming.ReceiverID, deleteEvent)
		}

	default:
		_ = client.WriteJSON(WSErrorEvent{
			Type:    "error",
			Message: "Unknown message type",
		})
	}
}

func (h *WSHandler) broadcastUserStatus(userID uint, isOnline bool, lastSeen *time.Time) {
	friends, err := h.friendService.GetFriends(userID)
	if err == nil {
		event := fiber.Map{
			"type":      "user_status",
			"user_id":   userID,
			"is_online": isOnline,
			"last_seen": lastSeen,
		}
		for _, f := range friends {
			h.hub.SendToUser(f.ID, event)
		}
	}
}
