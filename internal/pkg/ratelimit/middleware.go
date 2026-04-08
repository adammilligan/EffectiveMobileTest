package ratelimit

import (
	"net"
	"net/http"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/config"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/httpjson"
)

type entry struct {
	bucket   bucket
	lastSeen time.Time
}

type store struct {
	mu          sync.Mutex
	byKey       map[string]entry
	lastCleanup time.Time
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
	ratePerSec float64
	burst      float64
}

func newBucket(now time.Time, requestsPerMinute int, burst int) bucket {
	ratePerSec := float64(requestsPerMinute) / 60.0
	if ratePerSec <= 0 {
		ratePerSec = 1
	}

	burstF := float64(burst)
	if burstF <= 0 {
		burstF = 1
	}

	return bucket{
		tokens:     burstF,
		lastRefill: now,
		ratePerSec: ratePerSec,
		burst:      burstF,
	}
}

func (b *bucket) allow(now time.Time) (bool, time.Duration) {
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.ratePerSec
		if b.tokens > b.burst {
			b.tokens = b.burst
		}

		b.lastRefill = now
	}

	if b.tokens < 1 {
		waitSeconds := (1 - b.tokens) / b.ratePerSec
		if waitSeconds < 0 {
			waitSeconds = 0
		}

		return false, time.Duration(waitSeconds * float64(time.Second))
	}

	b.tokens -= 1

	// If the bucket is now empty (<1 token), compute when the next token will be available.
	if b.tokens < 1 {
		waitSeconds := (1 - b.tokens) / b.ratePerSec
		if waitSeconds < 0 {
			waitSeconds = 0
		}

		return true, time.Duration(waitSeconds * float64(time.Second))
	}

	return true, 0
}

func newStore() *store {
	return &store{
		byKey: make(map[string]entry),
	}
}

func (s *store) allowWithDetails(key string, requestsPerMinute int, burst int, now time.Time) (bool, time.Duration, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, isOk := s.byKey[key]; isOk {
		existing.lastSeen = now
		isAllowed, wait := existing.bucket.allow(now)
		remaining := int(math.Floor(existing.bucket.tokens))
		if remaining < 0 {
			remaining = 0
		}
		if remaining > burst {
			remaining = burst
		}
		s.byKey[key] = existing
		return isAllowed, wait, remaining
	}

	b := newBucket(now, requestsPerMinute, burst)
	isAllowed, wait := b.allow(now)
	remaining := int(math.Floor(b.tokens))
	if remaining < 0 {
		remaining = 0
	}
	if remaining > burst {
		remaining = burst
	}

	s.byKey[key] = entry{
		bucket:   b,
		lastSeen: now,
	}

	return isAllowed, wait, remaining
}

func (s *store) cleanup(now time.Time, maxIdle time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// avoid scanning map too often
	if !s.lastCleanup.IsZero() && now.Sub(s.lastCleanup) < time.Minute {
		return
	}

	for key, e := range s.byKey {
		if now.Sub(e.lastSeen) > maxIdle {
			delete(s.byKey, key)
		}
	}

	s.lastCleanup = now
}

func Middleware(cfg config.RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.IsEnabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	requestsPerMinute := cfg.RequestsPerMinute
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}

	burst := cfg.Burst
	if burst <= 0 {
		burst = 10
	}

	st := newStore()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			now := time.Now()
			st.cleanup(now, 10*time.Minute)

			ip := clientIP(req.RemoteAddr)
			allowed, wait, remaining := st.allowWithDetails(ip, requestsPerMinute, burst, now)

			// Provide rate-limit diagnostics for clients.
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			resetUnix := now.Add(wait).Unix()
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetUnix, 10))

			if !allowed {
				retryAfterSeconds := int(math.Ceil(wait.Seconds()))
				if retryAfterSeconds < 1 {
					retryAfterSeconds = 1
				}

				w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
				httpjson.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil && host != "" {
		return host
	}

	return remoteAddr
}

