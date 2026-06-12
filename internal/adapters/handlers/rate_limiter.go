package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipEntry struct {
	count   int
	resetAt time.Time
}

// RateLimiter es un rate limiter de ventana fija por IP.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
	max     int
	window  time.Duration
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*ipEntry),
		max:     max,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(10 * time.Minute)
		rl.mu.Lock()
		now := time.Now()
		for ip, e := range rl.entries {
			if now.After(e.resetAt) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		rl.mu.Lock()
		e, ok := rl.entries[ip]
		if !ok || now.After(e.resetAt) {
			rl.entries[ip] = &ipEntry{count: 1, resetAt: now.Add(rl.window)}
			rl.mu.Unlock()
			c.Next()
			return
		}
		e.count++
		if e.count > rl.max {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "demasiados intentos, esperá unos minutos"})
			c.Abort()
			return
		}
		rl.mu.Unlock()
		c.Next()
	}
}
