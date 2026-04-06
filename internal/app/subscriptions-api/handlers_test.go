package subscriptionsapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/subscriptions"
)

type fakeRepo struct{}

func (f fakeRepo) Create(ctx context.Context, p subscriptions.CreateParams) (subscriptions.Subscription, error) {
	return subscriptions.Subscription{}, nil
}
func (f fakeRepo) Get(ctx context.Context, id string) (subscriptions.Subscription, bool, error) {
	return subscriptions.Subscription{}, false, nil
}
func (f fakeRepo) Update(ctx context.Context, id string, p subscriptions.UpdateParams) (subscriptions.Subscription, bool, error) {
	return subscriptions.Subscription{}, false, nil
}
func (f fakeRepo) Delete(ctx context.Context, id string) (bool, error) { return false, nil }
func (f fakeRepo) List(ctx context.Context, p subscriptions.ListParams) ([]subscriptions.Subscription, error) {
	return nil, nil
}
func (f fakeRepo) FindOverlappingForTotal(ctx context.Context, p subscriptions.TotalQueryParams) ([]subscriptions.Subscription, error) {
	return nil, nil
}

func TestCreateValidation_EndDateBeforeStartDate(t *testing.T) {
	t.Parallel()

	h := NewHandlers(fakeRepo{})
	r := NewRouter(h)

	reqBody := []byte(`{"service_name":"X","price":1,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"08-2025","end_date":"07-2025"}`)
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

