package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"file-sharing-platform/internal/auth"
	"file-sharing-platform/pkg/cache"
)

type RateLimiter struct {
	cache         cache.Cache
	maxRequests   int
	windowSeconds int
}

func NewRateLimiter(cache cache.Cache, maxRequests, windowSeconds int) *RateLimiter {
	return &RateLimiter{
		cache:         cache,
		maxRequests:   maxRequests,
		windowSeconds: windowSeconds,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID or use IP address as fallback
		var identifier string
		userID, err := auth.GetUserIDFromRequest(r)
		if err == nil {
			identifier = strconv.FormatInt(userID, 10)
		} else {
			identifier = r.RemoteAddr
		}

		// Check rate limit
		key := "ratelimit:" + identifier

		// Get current count
		var countStr string
		err = rl.cache.Get(context.Background(), key, &countStr)
		var count int
		if err == nil {
			count, _ = strconv.Atoi(countStr)
		}

		// Check if limit exceeded
		if count >= rl.maxRequests {
			w.Header().Set("Retry-After", strconv.Itoa(rl.windowSeconds))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Increment counter
		expiration := time.Duration(rl.windowSeconds) * time.Second
		if count == 0 {
			// First request in window, set expiration
			rl.cache.Set(context.Background(), key, strconv.Itoa(count+1), expiration)
		} else {
			// Increment existing counter
			rl.cache.Set(context.Background(), key, strconv.Itoa(count+1), expiration)
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}
