package subscriptionsapi

import (
	"context"
	"errors"
	"testing"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/subscriptions"
)

type fakeRepoForService struct {
	createResult subscriptions.Subscription
	createError  error

	getResult   subscriptions.Subscription
	getIsFound  bool
	getError    error
	updateResult subscriptions.Subscription
	updateIsFound bool
	updateError   error

	deleteIsDeleted bool
	deleteError     error

	listResult []subscriptions.Subscription
	listError  error

	totalItems []subscriptions.Subscription
	totalError error
}

func (f *fakeRepoForService) Create(ctx context.Context, params subscriptions.CreateParams) (subscriptions.Subscription, error) {
	return f.createResult, f.createError
}

func (f *fakeRepoForService) Get(ctx context.Context, id string) (subscriptions.Subscription, bool, error) {
	return f.getResult, f.getIsFound, f.getError
}

func (f *fakeRepoForService) Update(ctx context.Context, id string, params subscriptions.UpdateParams) (subscriptions.Subscription, bool, error) {
	return f.updateResult, f.updateIsFound, f.updateError
}

func (f *fakeRepoForService) Delete(ctx context.Context, id string) (bool, error) {
	return f.deleteIsDeleted, f.deleteError
}

func (f *fakeRepoForService) List(ctx context.Context, params subscriptions.ListParams) ([]subscriptions.Subscription, error) {
	return f.listResult, f.listError
}

func (f *fakeRepoForService) FindOverlappingForTotal(ctx context.Context, params subscriptions.TotalQueryParams) ([]subscriptions.Subscription, error) {
	return f.totalItems, f.totalError
}

func TestServiceCreate_DateRangeValidation(t *testing.T) {
	t.Parallel()

	start, parseErr := subscriptions.ParseMonth("08-2025")
	if parseErr != nil {
		t.Fatalf("parse start: %v", parseErr)
	}

	endBefore, parseErr := subscriptions.ParseMonth("07-2025")
	if parseErr != nil {
		t.Fatalf("parse end: %v", parseErr)
	}

	repo := &fakeRepoForService{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), subscriptions.CreateParams{
		ServiceName: "X",
		PriceRub:    100,
		UserID:      "user",
		StartMonth:  start,
		EndMonth:    &endBefore,
	})

	if !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("want ErrInvalidDateRange, got %v", err)
	}
}

func TestServicePatch_DateRangeValidation(t *testing.T) {
	t.Parallel()

	start, parseErr := subscriptions.ParseMonth("08-2025")
	if parseErr != nil {
		t.Fatalf("parse start: %v", parseErr)
	}

	endBefore, parseErr := subscriptions.ParseMonth("07-2025")
	if parseErr != nil {
		t.Fatalf("parse end: %v", parseErr)
	}

	startCopy := start
	endCopy := endBefore

	params := subscriptions.UpdateParams{
		StartMonth: &startCopy,
		EndMonth:   &([]*subscriptions.Month{&endCopy}[0]),
	}

	repo := &fakeRepoForService{}
	svc := NewService(repo)

	_, _, err := svc.Patch(context.Background(), "id", params)
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("want ErrInvalidDateRange, got %v", err)
	}
}

func TestServiceTotal_InvalidPeriod(t *testing.T) {
	t.Parallel()

	from, parseErr := subscriptions.ParseMonth("09-2025")
	if parseErr != nil {
		t.Fatalf("parse from: %v", parseErr)
	}

	to, parseErr := subscriptions.ParseMonth("07-2025")
	if parseErr != nil {
		t.Fatalf("parse to: %v", parseErr)
	}

	repo := &fakeRepoForService{}
	svc := NewService(repo)

	_, err := svc.Total(context.Background(), subscriptions.TotalQueryParams{
		From: from,
		To:   to,
	})
	if !errors.Is(err, ErrInvalidPeriodDateRange) {
		t.Fatalf("want ErrInvalidPeriodDateRange, got %v", err)
	}
}

func TestServiceTotal_CalculatesTotal(t *testing.T) {
	t.Parallel()

	from, parseErr := subscriptions.ParseMonth("07-2025")
	if parseErr != nil {
		t.Fatalf("parse from: %v", parseErr)
	}

	to, parseErr := subscriptions.ParseMonth("09-2025")
	if parseErr != nil {
		t.Fatalf("parse to: %v", parseErr)
	}

	start, parseErr := subscriptions.ParseMonth("07-2025")
	if parseErr != nil {
		t.Fatalf("parse start: %v", parseErr)
	}

	end, parseErr := subscriptions.ParseMonth("08-2025")
	if parseErr != nil {
		t.Fatalf("parse end: %v", parseErr)
	}

	first := subscriptions.Subscription{
		PriceRub:   100,
		StartMonth: start,
		EndMonth:   &end,
	}

	openStart, parseErr := subscriptions.ParseMonth("09-2025")
	if parseErr != nil {
		t.Fatalf("parse open start: %v", parseErr)
	}

	second := subscriptions.Subscription{
		PriceRub:   200,
		StartMonth: openStart,
		EndMonth:   nil,
	}

	repo := &fakeRepoForService{
		totalItems: []subscriptions.Subscription{first, second},
	}

	svc := NewService(repo)

	total, err := svc.Total(context.Background(), subscriptions.TotalQueryParams{
		From: from,
		To:   to,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	// first: Jul–Aug (2 месяца) по 100 = 200
	// second: Sep (1 месяц в рамках периода) по 200 = 200
	// итого 400
	if total != 400 {
		t.Fatalf("want total 400, got %d", total)
	}
}

