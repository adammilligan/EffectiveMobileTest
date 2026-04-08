package subscriptionsapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/config"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/subscriptions"
)

type fakeServiceForRouterTest struct{}

func (f fakeServiceForRouterTest) Create(ctx context.Context, params subscriptions.CreateParams) (subscriptions.Subscription, error) {
	return subscriptions.Subscription{}, nil
}

func (f fakeServiceForRouterTest) Get(ctx context.Context, id string) (subscriptions.Subscription, bool, error) {
	return subscriptions.Subscription{}, false, nil
}

func (f fakeServiceForRouterTest) Patch(ctx context.Context, id string, params subscriptions.UpdateParams) (subscriptions.Subscription, bool, error) {
	return subscriptions.Subscription{}, false, nil
}

func (f fakeServiceForRouterTest) Delete(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (f fakeServiceForRouterTest) List(ctx context.Context, params subscriptions.ListParams) ([]subscriptions.Subscription, error) {
	return nil, nil
}

func (f fakeServiceForRouterTest) Total(ctx context.Context, params subscriptions.TotalQueryParams) (int, error) {
	return 0, nil
}

func TestRouter_RateLimit_AppliesToSubscriptionsOnly(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		RateLimit: config.RateLimitConfig{
			IsEnabled:         true,
			RequestsPerMinute: 1,
			Burst:             1,
		},
	}

	h := NewHandlers(fakeServiceForRouterTest{})
	r := NewRouter(cfg, h)

	t.Run("/healthz is not limited", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			req.RemoteAddr = "1.2.3.4:1111"
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("healthz call #%d: want %d, got %d", i+1, http.StatusOK, rr.Code)
			}
		}
	})

	t.Run("/subscriptions is limited", func(t *testing.T) {
		t.Parallel()

		body := []byte(`{"service_name":"X","price":1,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"08-2025"}`)

		req1 := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewReader(body))
		req1.Header.Set("Content-Type", "application/json")
		req1.RemoteAddr = "1.2.3.4:1111"
		rr1 := httptest.NewRecorder()
		r.ServeHTTP(rr1, req1)
		if rr1.Code == http.StatusTooManyRequests {
			t.Fatalf("expected first subscriptions request to pass, got %d", rr1.Code)
		}

		req2 := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")
		req2.RemoteAddr = "1.2.3.4:2222"
		rr2 := httptest.NewRecorder()
		r.ServeHTTP(rr2, req2)
		if rr2.Code != http.StatusTooManyRequests {
			t.Fatalf("want %d, got %d", http.StatusTooManyRequests, rr2.Code)
		}
		if rr2.Header().Get("Retry-After") == "" {
			t.Fatalf("expected Retry-After header")
		}
	})
}

