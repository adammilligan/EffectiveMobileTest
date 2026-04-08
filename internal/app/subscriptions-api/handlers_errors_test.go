package subscriptionsapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/config"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/subscriptions"
)

type errorService struct {
	createErr error
	getErr    error
	patchErr  error
	deleteErr error
	listErr   error
	totalErr  error

	getIsFound   bool
	patchIsFound bool
	deleteOK     bool
}

func (s errorService) Create(ctx context.Context, params subscriptions.CreateParams) (subscriptions.Subscription, error) {
	return subscriptions.Subscription{}, s.createErr
}

func (s errorService) Get(ctx context.Context, id string) (subscriptions.Subscription, bool, error) {
	return subscriptions.Subscription{}, s.getIsFound, s.getErr
}

func (s errorService) Patch(ctx context.Context, id string, params subscriptions.UpdateParams) (subscriptions.Subscription, bool, error) {
	return subscriptions.Subscription{}, s.patchIsFound, s.patchErr
}

func (s errorService) Delete(ctx context.Context, id string) (bool, error) {
	return s.deleteOK, s.deleteErr
}

func (s errorService) List(ctx context.Context, params subscriptions.ListParams) ([]subscriptions.Subscription, error) {
	return nil, s.listErr
}

func (s errorService) Total(ctx context.Context, params subscriptions.TotalQueryParams) (int, error) {
	return 0, s.totalErr
}

func TestHandlers_Get_Returns404WhenNotFound(t *testing.T) {
	t.Parallel()

	h := NewHandlers(errorService{getIsFound: false})
	r := NewRouter(config.Config{RateLimit: config.RateLimitConfig{IsEnabled: false}}, h)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/60601fee-2bf1-4721-ae6f-7636e79a0cba", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("want %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

func TestHandlers_Delete_Returns404WhenNotFound(t *testing.T) {
	t.Parallel()

	h := NewHandlers(errorService{deleteOK: false})
	r := NewRouter(config.Config{RateLimit: config.RateLimitConfig{IsEnabled: false}}, h)

	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/60601fee-2bf1-4721-ae6f-7636e79a0cba", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("want %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

func TestHandlers_List_Returns400OnBadUserID(t *testing.T) {
	t.Parallel()

	h := NewHandlers(errorService{})
	r := NewRouter(config.Config{RateLimit: config.RateLimitConfig{IsEnabled: false}}, h)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions?user_id=not-a-uuid", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestHandlers_Total_Returns400OnBadFromTo(t *testing.T) {
	t.Parallel()

	h := NewHandlers(errorService{})
	r := NewRouter(config.Config{RateLimit: config.RateLimitConfig{IsEnabled: false}}, h)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=2025-07&to=09-2025", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=07-2025&to=2025-09", nil)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("want %d, got %d, body=%s", http.StatusBadRequest, rr2.Code, rr2.Body.String())
	}
}

func TestHandlers_Total_Returns500OnServiceError(t *testing.T) {
	t.Parallel()

	h := NewHandlers(errorService{totalErr: errors.New("boom")})
	r := NewRouter(config.Config{RateLimit: config.RateLimitConfig{IsEnabled: false}}, h)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=07-2025&to=09-2025", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("want %d, got %d, body=%s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
}

func TestHandlers_Patch_Returns500OnServiceError(t *testing.T) {
	t.Parallel()

	h := NewHandlers(errorService{patchErr: errors.New("boom"), patchIsFound: true})
	r := NewRouter(config.Config{RateLimit: config.RateLimitConfig{IsEnabled: false}}, h)

	body := []byte(`{"service_name":"X"}`)
	req := httptest.NewRequest(http.MethodPatch, "/subscriptions/60601fee-2bf1-4721-ae6f-7636e79a0cba", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("want %d, got %d, body=%s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
}

func TestHandlers_Patch_Returns404WhenNotFound(t *testing.T) {
	t.Parallel()

	h := NewHandlers(errorService{patchIsFound: false})
	r := NewRouter(config.Config{RateLimit: config.RateLimitConfig{IsEnabled: false}}, h)

	body := []byte(`{"service_name":"X"}`)
	req := httptest.NewRequest(http.MethodPatch, "/subscriptions/60601fee-2bf1-4721-ae6f-7636e79a0cba", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("want %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

