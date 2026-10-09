package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	visitorIdleFor = 10 * time.Minute
	cleanupEvery   = 1 * time.Minute
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPLimiter struct {
	mu       sync.Mutex
	limiters map[string]*visitor
	r        rate.Limit
	burst    int
}

func (s *IPLimiter) cleanupLoop() {
	ticker := time.NewTicker(cleanupEvery)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for ip, v := range s.limiters {
			if time.Since(v.lastSeen) > visitorIdleFor {
				delete(s.limiters, ip)
			}
		}
		s.mu.Unlock()
	}
}

func NewIPLimiter(r rate.Limit, burst int) *IPLimiter {
	s := &IPLimiter{
		limiters: make(map[string]*visitor),
		r:        r,
		burst:    burst,
	}
	go s.cleanupLoop()
	return s
}

func (s *IPLimiter) get(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.limiters[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(s.r, s.burst)}
		s.limiters[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (s *IPLimiter) RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !s.get(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests,
				gin.H{
					"status":  false,
					"message": "Too many requests",
				})
			return
		}
		c.Next()
	}
}
