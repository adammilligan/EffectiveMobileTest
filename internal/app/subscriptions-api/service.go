package subscriptionsapi

import (
	"context"
	"errors"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/subscriptions"
)

var (
	ErrInvalidDateRange       = errors.New("end_date must be >= start_date")
	ErrInvalidPeriodDateRange = errors.New("from must be <= to")
)

type Service interface {
	Create(ctx context.Context, params subscriptions.CreateParams) (subscriptions.Subscription, error)
	Get(ctx context.Context, id string) (subscriptions.Subscription, bool, error)
	Patch(ctx context.Context, id string, params subscriptions.UpdateParams) (subscriptions.Subscription, bool, error)
	Delete(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, params subscriptions.ListParams) ([]subscriptions.Subscription, error)
	Total(ctx context.Context, params subscriptions.TotalQueryParams) (int, error)
}

type service struct {
	repo repo
}

func NewService(repo repo) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, params subscriptions.CreateParams) (subscriptions.Subscription, error) {
	if params.EndMonth != nil && params.EndMonth.IsBefore(params.StartMonth) {
		return subscriptions.Subscription{}, ErrInvalidDateRange
	}

	created, err := s.repo.Create(ctx, params)
	if err != nil {
		return subscriptions.Subscription{}, err
	}

	return created, nil
}

func (s *service) Get(ctx context.Context, id string) (subscriptions.Subscription, bool, error) {
	subscription, isFound, err := s.repo.Get(ctx, id)
	if err != nil {
		return subscriptions.Subscription{}, false, err
	}

	if !isFound {
		return subscriptions.Subscription{}, false, nil
	}

	return subscription, true, nil
}

func (s *service) Patch(ctx context.Context, id string, params subscriptions.UpdateParams) (subscriptions.Subscription, bool, error) {
	if params.StartMonth != nil && params.EndMonth != nil && *params.EndMonth != nil && (**params.EndMonth).IsBefore(*params.StartMonth) {
		return subscriptions.Subscription{}, false, ErrInvalidDateRange
	}

	updated, isFound, err := s.repo.Update(ctx, id, params)
	if err != nil {
		return subscriptions.Subscription{}, false, err
	}

	if !isFound {
		return subscriptions.Subscription{}, false, nil
	}

	return updated, true, nil
}

func (s *service) Delete(ctx context.Context, id string) (bool, error) {
	isDeleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return false, err
	}

	return isDeleted, nil
}

func (s *service) List(ctx context.Context, params subscriptions.ListParams) ([]subscriptions.Subscription, error) {
	items, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *service) Total(ctx context.Context, params subscriptions.TotalQueryParams) (int, error) {
	if params.From.Time().After(params.To.Time()) {
		return 0, ErrInvalidPeriodDateRange
	}

	items, err := s.repo.FindOverlappingForTotal(ctx, params)
	if err != nil {
		return 0, err
	}

	total := 0

	for _, item := range items {
		months := subscriptions.OverlapMonthsInclusive(item.StartMonth, item.EndMonth, params.From, params.To)
		total += months * item.PriceRub
	}

	return total, nil
}

