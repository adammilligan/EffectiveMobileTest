package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/config"
)

func TestMiddleware_Returns429AfterBurst(t *testing.T) {
	t.Parallel()

	mw := Middleware(config.RateLimitConfig{
		IsEnabled:         true,
		RequestsPerMinute: 1,
		Burst:             1,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	req1.RemoteAddr = "1.2.3.4:1111"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	req2.RemoteAddr = "1.2.3.4:2222"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("want %d, got %d", http.StatusTooManyRequests, rr2.Code)
	}

	if got := rr2.Header().Get("Retry-After"); got == "" {
		t.Fatalf("expected Retry-After header")
	}

	if rr2.Header().Get("X-RateLimit-Limit") == "" {
		t.Fatalf("expected X-RateLimit-Limit header")
	}
	if rr2.Header().Get("X-RateLimit-Remaining") == "" {
		t.Fatalf("expected X-RateLimit-Remaining header")
	}
	if rr2.Header().Get("X-RateLimit-Reset") == "" {
		t.Fatalf("expected X-RateLimit-Reset header")
	}
}

func TestMiddleware_DifferentIPsHaveSeparateBuckets(t *testing.T) {
	t.Parallel()

	mw := Middleware(config.RateLimitConfig{
		IsEnabled:         true,
		RequestsPerMinute: 1,
		Burst:             1,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	req1.RemoteAddr = "1.1.1.1:1111"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	req2.RemoteAddr = "2.2.2.2:2222"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, rr2.Code)
	}
}

func TestMiddleware_RemoteAddrWithoutPort(t *testing.T) {
	t.Parallel()

	mw := Middleware(config.RateLimitConfig{
		IsEnabled:         true,
		RequestsPerMinute: 1,
		Burst:             1,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	req1.RemoteAddr = "1.2.3.4"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	req2.RemoteAddr = "1.2.3.4"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("want %d, got %d", http.StatusTooManyRequests, rr2.Code)
	}
}

