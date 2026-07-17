package realtime

import (
	"sync"
	"time"
)

const (
	sessionTTL = 5 * time.Minute
	sweepEvery = 30 * time.Second
)

// ActiveSessions rastrea sesiones activas en memoria (session_id → last_seen),
// protegido con RWMutex.
type ActiveSessions struct {
	mu       sync.RWMutex
	lastSeen map[string]time.Time
}

// NewActiveSessions crea el store y lanza el janitor en background que descarta
// sesiones inactivas (> 5 min) cada 30s.
func NewActiveSessions() *ActiveSessions {
	a := &ActiveSessions{lastSeen: make(map[string]time.Time)}
	go a.janitor()
	return a
}

// Touch marca la sesión como vista recién. Ignora session_id vacío.
func (a *ActiveSessions) Touch(sessionID string) {
	if sessionID == "" {
		return
	}
	a.mu.Lock()
	a.lastSeen[sessionID] = time.Now()
	a.mu.Unlock()
}

// GetActiveSessionCount devuelve la cantidad de sesiones activas.
func (a *ActiveSessions) GetActiveSessionCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.lastSeen)
}

func (a *ActiveSessions) janitor() {
	ticker := time.NewTicker(sweepEvery)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-sessionTTL)
		a.mu.Lock()
		for id, seen := range a.lastSeen {
			if seen.Before(cutoff) {
				delete(a.lastSeen, id)
			}
		}
		a.mu.Unlock()
	}
}
