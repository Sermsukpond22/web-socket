package websocket

import (
	"chat_app/backend/models"
	"sync"
)

type RoomRepository interface {
	GetJoinedRoomMembers(roomID uint) ([]models.RoomMember, error)
}

type Hub struct {
	clients  map[uint][]*Client
	mu       sync.RWMutex
	roomRepo RoomRepository
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uint][]*Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.UserID] = append(h.clients[client.UserID], client)
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.clients[client.UserID]
	for i, c := range clients {
		if c == client {
			h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(h.clients[client.UserID]) == 0 {
		delete(h.clients, client.UserID)
	}
	client.Close()
}

func (h *Hub) GetClients(userID uint) ([]*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userID]
	return clients, ok
}

func (h *Hub) SendToUser(userID uint, payload interface{}) bool {
	clients, ok := h.GetClients(userID)
	if !ok || len(clients) == 0 {
		return false
	}

	success := false
	for _, client := range clients {
		err := client.WriteJSON(payload)
		if err != nil {
			// We can't unregister while iterating like this because it calls lock, but Unregister uses Lock.
			// Actually SendToUser uses GetClients which only uses RLock, so calling Unregister (which uses Lock) is safe here
			// because GetClients already released RLock. Let's just launch a goroutine to unregister to be safe and avoid deadlocks if the client is slow.
			go h.Unregister(client)
		} else {
			success = true
		}
	}
	return success
}

func (h *Hub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

func (h *Hub) SetRoomRepo(repo RoomRepository) {
	h.roomRepo = repo
}

func (h *Hub) SendToRoom(roomID uint, payload interface{}) bool {
	if h.roomRepo == nil {
		return false
	}
	members, err := h.roomRepo.GetJoinedRoomMembers(roomID)
	if err != nil {
		return false
	}

	success := false
	for _, member := range members {
		if h.SendToUser(member.UserID, payload) {
			success = true
		}
	}
	return success
}
