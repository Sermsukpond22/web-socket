package websocket_test

import (
	"sync"
	"testing"

	wsPkg "chat_app/backend/websocket"
)

func TestHub_ThreadSafety(t *testing.T) {
	hub := wsPkg.NewHub()

	var wg sync.WaitGroup
	for i := uint(1); i <= 50; i++ {
		wg.Add(1)
		go func(id uint) {
			defer wg.Done()
			client := wsPkg.NewClient(id, nil, hub)
			hub.Register(client)
			if !hub.IsUserOnline(id) {
				t.Errorf("expected user %d to be online", id)
			}
			hub.Unregister(client)
			if hub.IsUserOnline(id) {
				t.Errorf("expected user %d to be offline", id)
			}
		}(i)
	}
	wg.Wait()
}
