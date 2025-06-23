package utils

import (
	"sync"
	"time"
)

// RateLimiter implémente un limiteur de débit simple
type RateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	ips    map[string][]time.Time
}

// NewRateLimiter crée une nouvelle instance de RateLimiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:  limit,
		window: window,
		ips:    make(map[string][]time.Time),
	}
}

// Allow vérifie si une requête est autorisée
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Nettoyage des anciennes requêtes
	if timestamps, exists := rl.ips[ip]; exists {
		var validTimestamps []time.Time
		for _, ts := range timestamps {
			if ts.After(windowStart) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		rl.ips[ip] = validTimestamps
	}

	// Vérification de la limite
	if timestamps, exists := rl.ips[ip]; exists {
		if len(timestamps) >= rl.limit {
			return false
		}
		rl.ips[ip] = append(timestamps, now)
	} else {
		rl.ips[ip] = []time.Time{now}
	}

	return true
} 