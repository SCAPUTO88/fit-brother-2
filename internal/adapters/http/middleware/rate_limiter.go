package middleware

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count    int
	resetAt  time.Time
	lastSeen time.Time
}

type RateLimiter struct {
	mu              sync.Mutex
	entries         map[string]*rateLimitEntry
	limit           int
	window          time.Duration
	cleanupInterval time.Duration
	lastSweep       time.Time
	nowFunc         func() time.Time
}

func NewRateLimiter(limit int, window time.Duration, cleanupInterval time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 20
	}
	if window <= 0 {
		window = time.Minute
	}
	if cleanupInterval <= 0 {
		cleanupInterval = 5 * time.Minute
	}

	now := time.Now()
	return &RateLimiter{
		entries:         make(map[string]*rateLimitEntry),
		limit:           limit,
		window:          window,
		cleanupInterval: cleanupInterval,
		lastSweep:       now,
		nowFunc:         time.Now,
	}
}

func (rl *RateLimiter) Allow(key string) (bool, time.Duration) {
	now := rl.nowFunc()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if now.Sub(rl.lastSweep) >= rl.cleanupInterval {
		rl.sweepExpiredLocked(now)
		rl.lastSweep = now
	}

	entry, exists := rl.entries[key]
	if !exists || !now.Before(entry.resetAt) {
		rl.entries[key] = &rateLimitEntry{
			count:    1,
			resetAt:  now.Add(rl.window),
			lastSeen: now,
		}
		return true, 0
	}

	entry.lastSeen = now
	if entry.count < rl.limit {
		entry.count++
		return true, 0
	}

	return false, entry.resetAt.Sub(now)
}

func (rl *RateLimiter) sweepExpiredLocked(now time.Time) {
	for k, e := range rl.entries {
		if !now.Before(e.resetAt) && now.Sub(e.lastSeen) >= rl.window {
			delete(rl.entries, k)
		}
	}
}

func (rl *RateLimiter) Middleware(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := strings.TrimSpace(c.ClientIP())
		if clientIP == "" || clientIP == "::1" {
			clientIP = "127.0.0.1"
		}
		key := scope + ":" + clientIP

		allowed, retryAfter := rl.Allow(key)
		if !allowed {
			secs := int(math.Ceil(retryAfter.Seconds()))
			if secs < 1 {
				secs = 1
			}
			c.Header("Retry-After", strconv.Itoa(secs))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "muitas tentativas; aguarde antes de tentar novamente",
			})
			return
		}

		c.Next()
	}
}
