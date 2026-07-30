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
	ReceiverID  uint   `json:"receiver_id"`
	RecipientID uint   `json:"recipient_id"`
	Content     string `json:"content"`
	IsTyping    *bool  `json:"is_typing,omitempty"`
}

type WSChatEvent struct {
	Type       string    `json:"type"`
	ID         uint      `json:"id"`
	SenderID   uint      `json:"sender_id"`
	ReceiverID uint      `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type WSErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type MessageService interface {
	SendMessage(senderID, receiverID uint, content string) (*models.Message, error)
}

type FriendService interface {
	AreFriends(userID1, userID2 uint) (bool, error)
}

type WSHandler struct {
	hub            *Hub
	messageService MessageService
	friendService  FriendService
}

func NewWSHandler(hub *Hub, messageService MessageService, friendService FriendService) *WSHandler {
	return &WSHandler{
		hub:            hub,
		messageService: messageService,
		friendService:  friendService,
	}
}

func (h *WSHandler) HandleConnection(c *websocket.Conn) {
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
	defer h.hub.Unregister(client)

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
		receiverID := incoming.ReceiverID
		if receiverID == 0 {
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
		savedMsg, err := h.messageService.SendMessage(client.UserID, incoming.ReceiverID, incoming.Content)
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
			Content:    savedMsg.Content,
			CreatedAt:  savedMsg.CreatedAt,
		}

		// Confirm back to sender
		_ = client.WriteJSON(chatEvent)

		// Send to target receiver if online (and not same client)
		if savedMsg.ReceiverID != client.UserID {
			h.hub.SendToUser(savedMsg.ReceiverID, chatEvent)
		}
	default:
		_ = client.WriteJSON(WSErrorEvent{
			Type:    "error",
			Message: "Unknown message type",
		})
	}
}
