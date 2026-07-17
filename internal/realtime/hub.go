package realtime

import "sync"

// Hub es un pub/sub en memoria para difundir mensajes a los streams SSE activos.
type Hub struct {
	mu   sync.RWMutex
	subs map[chan string]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[chan string]struct{})}
}

// Subscribe registra un nuevo suscriptor y devuelve su canal (buffer 16).
func (h *Hub) Subscribe() chan string {
	ch := make(chan string, 16)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Unsubscribe da de baja y cierra el canal.
func (h *Hub) Unsubscribe(ch chan string) {
	h.mu.Lock()
	if _, ok := h.subs[ch]; ok {
		delete(h.subs, ch)
		close(ch)
	}
	h.mu.Unlock()
}

// Broadcast envía msg a todos los suscriptores. Si un suscriptor está lento
// (buffer lleno), descarta el mensaje para no bloquear al resto.
func (h *Hub) Broadcast(msg string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs {
		select {
		case ch <- msg:
		default:
		}
	}
}
