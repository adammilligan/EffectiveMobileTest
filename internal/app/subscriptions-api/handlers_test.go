package subscriptionsapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/subscriptions"
)

type fakeService struct{}

func (f fakeService) Create(ctx context.Context, params subscriptions.CreateParams) (subscriptions.Subscription, error) {
	if params.EndMonth != nil && params.EndMonth.IsBefore(params.StartMonth) {
		return subscriptions.Subscription{}, ErrInvalidDateRange
	}

	return subscriptions.Subscription{}, nil
}

func (f fakeService) Get(ctx context.Context, id string) (subscriptions.Subscription, bool, error) {
	return subscriptions.Subscription{}, false, nil
}

func (f fakeService) Patch(ctx context.Context, id string, params subscriptions.UpdateParams) (subscriptions.Subscription, bool, error) {
	return subscriptions.Subscription{}, false, nil
}

func (f fakeService) Delete(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (f fakeService) List(ctx context.Context, params subscriptions.ListParams) ([]subscriptions.Subscription, error) {
	return nil, nil
}

func (f fakeService) Total(ctx context.Context, params subscriptions.TotalQueryParams) (int, error) {
	return 0, nil
}

func TestCreateValidation(t *testing.T) {
	t.Parallel()

	h := NewHandlers(fakeService{})
	r := NewRouter(h)

	tests := map[string]struct {
		body       string
		wantStatus int
	}{
		"invalid json": {
			body:       `{"service_name":`,
			wantStatus: http.StatusBadRequest,
		},
		"empty service_name": {
			body:       `{"service_name":"","price":1,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"08-2025"}`,
			wantStatus: http.StatusBadRequest,
		},
		"negative price": {
			body:       `{"service_name":"X","price":-1,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"08-2025"}`,
			wantStatus: http.StatusBadRequest,
		},
		"bad user_id": {
			body:       `{"service_name":"X","price":1,"user_id":"not-a-uuid","start_date":"08-2025"}`,
			wantStatus: http.StatusBadRequest,
		},
		"bad start_date": {
			body:       `{"service_name":"X","price":1,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"2025-08"}`,
			wantStatus: http.StatusBadRequest,
		},
		"bad end_date format": {
			body:       `{"service_name":"X","price":1,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"08-2025","end_date":"2025-09"}`,
			wantStatus: http.StatusBadRequest,
		},
		"end_date before start_date": {
			body:       `{"service_name":"X","price":1,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"08-2025","end_date":"07-2025"}`,
			wantStatus: http.StatusBadRequest,
		},
		"ok": {
			body:       `{"service_name":"X","price":1,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"08-2025"}`,
			wantStatus: http.StatusCreated,
		},
	}

	for name, tc := range tests {
		name := name
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("want %d, got %d, body=%s", tc.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

