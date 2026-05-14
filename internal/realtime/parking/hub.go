package parking

import (
	"encoding/json"
	"log"
	"sync"
)

// Session là interface cho transport layer (WebTransport, WebSocket, etc.)
type Session interface {
	Send([]byte) error
	Close() error
}

// Client đại diện cho 1 frontend connection, thuộc về 1 lot.
type Client struct {
	LotID   uint64
	Session Session
}

// EventEnvelope là JSON envelope gửi cho frontend.
type EventEnvelope struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// ─── lotRoom ─────────────────────────────────────────────────────────────────
// Mỗi lotRoom chứa các clients đang xem cùng 1 lot.
// Lock riêng per-room → broadcast lot A không block lot B.
type lotRoom struct {
	mu      sync.RWMutex
	clients map[Session]*Client
}

func newRoom() *lotRoom {
	return &lotRoom{
		clients: make(map[Session]*Client),
	}
}

func (r *lotRoom) add(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[c.Session] = c
}

func (r *lotRoom) remove(s Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, s)
}

func (r *lotRoom) count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// snapshot trả về bản copy danh sách clients (tránh hold lock khi send).
func (r *lotRoom) snapshot() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*Client, 0, len(r.clients))
	for _, c := range r.clients {
		out = append(out, c)
	}
	return out
}

// ─── Hub ─────────────────────────────────────────────────────────────────────
// Hub quản lý tất cả rooms, sharded theo lotID.
// Mỗi room có lock riêng → giảm contention khi 10k+ clients.

const defaultFanOutWorkers = 64

type Hub struct {
	mu   sync.RWMutex
	lots map[uint64]*lotRoom

	fanOutWorkers int
}

func NewHub() *Hub {
	return &Hub{
		lots:          make(map[uint64]*lotRoom),
		fanOutWorkers: defaultFanOutWorkers,
	}
}

// getOrCreateRoom lấy hoặc tạo room cho lotID (thread-safe).
func (h *Hub) getOrCreateRoom(lotID uint64) *lotRoom {
	// Fast path: read lock
	h.mu.RLock()
	room, ok := h.lots[lotID]
	h.mu.RUnlock()
	if ok {
		return room
	}

	// Slow path: write lock
	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check
	if room, ok = h.lots[lotID]; ok {
		return room
	}

	room = newRoom()
	h.lots[lotID] = room
	return room
}

// getRoom lấy room nếu tồn tại (nil nếu không có).
func (h *Hub) getRoom(lotID uint64) *lotRoom {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lots[lotID]
}

// Add thêm client vào room tương ứng.
func (h *Hub) Add(client *Client) {
	room := h.getOrCreateRoom(client.LotID)
	room.add(client)
	log.Printf("[HUB] client added lotID=%d total_in_room=%d", client.LotID, room.count())
}

// Remove xóa client khỏi room. Cleanup room rỗng.
func (h *Hub) Remove(s Session) {
	h.mu.RLock()
	var targetLotID uint64
	var targetRoom *lotRoom
	for lotID, room := range h.lots {
		room.mu.RLock()
		if _, exists := room.clients[s]; exists {
			targetLotID = lotID
			targetRoom = room
		}
		room.mu.RUnlock()
		if targetRoom != nil {
			break
		}
	}
	h.mu.RUnlock()

	if targetRoom == nil {
		return
	}

	targetRoom.remove(s)

	// Cleanup room rỗng
	remaining := targetRoom.count()
	if remaining == 0 {
		h.mu.Lock()
		// Double-check: chỉ xóa nếu vẫn rỗng
		if targetRoom.count() == 0 {
			delete(h.lots, targetLotID)
		}
		h.mu.Unlock()
	}

	log.Printf("[HUB] client removed lotID=%d remaining=%d", targetLotID, remaining)
}

// BroadcastToLot gửi 1 event đến tất cả clients trong 1 lot.
// Dùng parallel fan-out với worker pool.
func (h *Hub) BroadcastToLot(lotID uint64, event string, data any) {
	room := h.getRoom(lotID)
	if room == nil {
		return
	}

	payload, err := json.Marshal(EventEnvelope{Event: event, Data: data})
	if err != nil {
		log.Printf("[HUB] marshal failed event=%s err=%v", event, err)
		return
	}

	h.fanOut(room, payload)
}

// BroadcastBatchToLot gửi batch events cho 1 lot.
// Tối ưu: marshal 1 lần, gửi 1 payload chứa tất cả events.
func (h *Hub) BroadcastBatchToLot(lotID uint64, event string, data any) {
	room := h.getRoom(lotID)
	if room == nil {
		return
	}

	payload, err := json.Marshal(EventEnvelope{Event: event, Data: data})
	if err != nil {
		log.Printf("[HUB] marshal batch failed event=%s err=%v", event, err)
		return
	}

	h.fanOut(room, payload)
}

// BroadcastAll gửi event cho tất cả clients ở mọi lot.
func (h *Hub) BroadcastAll(event string, data any) {
	payload, err := json.Marshal(EventEnvelope{Event: event, Data: data})
	if err != nil {
		log.Printf("[HUB] marshal failed event=%s err=%v", event, err)
		return
	}

	h.mu.RLock()
	rooms := make([]*lotRoom, 0, len(h.lots))
	for _, room := range h.lots {
		rooms = append(rooms, room)
	}
	h.mu.RUnlock()

	for _, room := range rooms {
		h.fanOut(room, payload)
	}
}

// fanOut gửi payload cho tất cả clients trong room sử dụng worker pool.
// Worker pool giới hạn concurrent goroutines (mặc định 64).
func (h *Hub) fanOut(room *lotRoom, payload []byte) {
	clients := room.snapshot()
	if len(clients) == 0 {
		return
	}

	// Nếu ít clients, gửi trực tiếp không cần pool
	if len(clients) <= h.fanOutWorkers {
		var wg sync.WaitGroup
		wg.Add(len(clients))
		for _, c := range clients {
			go func(client *Client) {
				defer wg.Done()
				if err := client.Session.Send(payload); err != nil {
					log.Printf("[HUB] send failed lotID=%d err=%v", client.LotID, err)
					h.Remove(client.Session)
					_ = client.Session.Close()
				}
			}(c)
		}
		wg.Wait()
		return
	}

	// Worker pool pattern cho nhiều clients
	jobs := make(chan *Client, len(clients))
	var wg sync.WaitGroup

	// Spawn workers
	for i := 0; i < h.fanOutWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for client := range jobs {
				if err := client.Session.Send(payload); err != nil {
					log.Printf("[HUB] send failed lotID=%d err=%v", client.LotID, err)
					h.Remove(client.Session)
					_ = client.Session.Close()
				}
			}
		}()
	}

	// Dispatch jobs
	for _, c := range clients {
		jobs <- c
	}
	close(jobs)

	wg.Wait()
}

// ClientCount trả về tổng số clients trên tất cả rooms.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, room := range h.lots {
		total += room.count()
	}
	return total
}

// RoomCount trả về số rooms (lots) đang active.
func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.lots)
}
