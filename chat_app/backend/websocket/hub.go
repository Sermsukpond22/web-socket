package websocket

import (
	"sync"
)

type Hub struct {
	clients map[uint]*Client
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uint]*Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if oldClient, ok := h.clients[client.UserID]; ok && oldClient != client {
		oldClient.Close()
	}
	h.clients[client.UserID] = client
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if currentClient, ok := h.clients[client.UserID]; ok && currentClient == client {
		delete(h.clients, client.UserID)
		client.Close()
	}
}

func (h *Hub) GetClient(userID uint) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[userID]
	return client, ok
}

func (h *Hub) SendToUser(userID uint, payload interface{}) bool {
	client, ok := h.GetClient(userID)
	if !ok {
		return false
	}
	err := client.WriteJSON(payload)
	if err != nil {
		h.Unregister(client)
		return false
	}
	return true
}

func (h *Hub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}
