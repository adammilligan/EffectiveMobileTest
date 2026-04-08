package subscriptionsapi

import (
	"context"
	"errors"
	"testing"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/subscriptions"
)

func TestService_PropagatesRepoErrors(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repo error")
	repo := &fakeRepoForService{
		createError: repoErr,
		getError:    repoErr,
		updateError: repoErr,
		deleteError: repoErr,
		listError:   repoErr,
		totalError:  repoErr,
	}

	svc := NewService(repo)

	start, parseErr := subscriptions.ParseMonth("07-2025")
	if parseErr != nil {
		t.Fatalf("parse start: %v", parseErr)
	}

	_, err := svc.Create(context.Background(), subscriptions.CreateParams{
		ServiceName: "X",
		PriceRub:    1,
		UserID:      "u",
		StartMonth:  start,
		EndMonth:    nil,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("Create: want %v, got %v", repoErr, err)
	}

	_, _, err = svc.Get(context.Background(), "id")
	if !errors.Is(err, repoErr) {
		t.Fatalf("Get: want %v, got %v", repoErr, err)
	}

	_, _, err = svc.Patch(context.Background(), "id", subscriptions.UpdateParams{})
	if !errors.Is(err, repoErr) {
		t.Fatalf("Patch: want %v, got %v", repoErr, err)
	}

	_, err = svc.Delete(context.Background(), "id")
	if !errors.Is(err, repoErr) {
		t.Fatalf("Delete: want %v, got %v", repoErr, err)
	}

	_, err = svc.List(context.Background(), subscriptions.ListParams{})
	if !errors.Is(err, repoErr) {
		t.Fatalf("List: want %v, got %v", repoErr, err)
	}

	_, err = svc.Total(context.Background(), subscriptions.TotalQueryParams{From: start, To: start})
	if !errors.Is(err, repoErr) {
		t.Fatalf("Total: want %v, got %v", repoErr, err)
	}
}

