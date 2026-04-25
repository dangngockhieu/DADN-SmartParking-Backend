package parking

import (
	"encoding/json"
	"sync"
)

type Hub struct {
	mu       sync.RWMutex
	sessions map[Session]struct{}
}

type Session interface {
	Send([]byte) error
	Close() error
}

type EventEnvelope struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func NewHub() *Hub {
	return &Hub{
		sessions: make(map[Session]struct{}),
	}
}

func (h *Hub) Add(s Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[s] = struct{}{}
}

func (h *Hub) Remove(s Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, s)
}

func (h *Hub) Broadcast(event string, data any) {
	payload, err := json.Marshal(EventEnvelope{
		Event: event,
		Data:  data,
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	sessions := make([]Session, 0, len(h.sessions))
	for s := range h.sessions {
		sessions = append(sessions, s)
	}
	h.mu.RUnlock()

	for _, s := range sessions {
		_ = s.Send(payload)
	}
}
