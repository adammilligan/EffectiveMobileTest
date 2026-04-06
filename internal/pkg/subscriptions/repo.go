package subscriptions

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateParams struct {
	ServiceName string
	PriceRub    int
	UserID      string
	StartMonth  Month
	EndMonth    *Month
}

type UpdateParams struct {
	ServiceName *string
	PriceRub    *int
	StartMonth  *Month
	EndMonth    **Month
}

type ListParams struct {
	UserID      *string
	ServiceName *string
	Limit       int
	Offset      int
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Create(ctx context.Context, p CreateParams) (Subscription, error) {
	const q = `
INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`

	start := p.StartMonth.Time()
	var end *time.Time
	if p.EndMonth != nil {
		t := p.EndMonth.Time()
		end = &t
	}

	row := r.pool.QueryRow(ctx, q, p.ServiceName, p.PriceRub, p.UserID, start, end)
	return scanSubscription(row)
}

func (r *Repo) Get(ctx context.Context, id string) (Subscription, bool, error) {
	const q = `
SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
FROM subscriptions
WHERE id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	s, err := scanSubscription(row)
	if err != nil {
		if isNoRows(err) {
			return Subscription{}, false, nil
		}
		return Subscription{}, false, err
	}
	return s, true, nil
}

func (r *Repo) Delete(ctx context.Context, id string) (bool, error) {
	const q = `DELETE FROM subscriptions WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

func (r *Repo) List(ctx context.Context, p ListParams) ([]Subscription, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	const q = `
SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
FROM subscriptions
WHERE ($1::uuid IS NULL OR user_id = $1::uuid)
  AND ($2::text IS NULL OR service_name = $2::text)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4`

	var userID *string = p.UserID
	var serviceName *string = p.ServiceName

	rows, err := r.pool.Query(ctx, q, userID, serviceName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Subscription
	for rows.Next() {
		s, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) Update(ctx context.Context, id string, p UpdateParams) (Subscription, bool, error) {
	// EndMonth uses **Month to distinguish:
	// - nil: not provided
	// - &nil: set end_date = NULL
	// - &m: set end_date = m
	const q = `
UPDATE subscriptions
SET
  service_name = COALESCE($2, service_name),
  price        = COALESCE($3, price),
  start_date   = COALESCE($4, start_date),
  end_date     = CASE
                  WHEN $5::date IS NULL AND $6::boolean IS TRUE THEN NULL
                  WHEN $5::date IS NOT NULL THEN $5::date
                  ELSE end_date
                END,
  updated_at   = NOW()
WHERE id = $1
RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`

	var start *time.Time
	if p.StartMonth != nil {
		t := p.StartMonth.Time()
		start = &t
	}

	var end *time.Time
	isEndProvided := false
	if p.EndMonth != nil {
		isEndProvided = true
		if *p.EndMonth != nil {
			t := (*p.EndMonth).Time()
			end = &t
		}
	}

	row := r.pool.QueryRow(ctx, q, id, p.ServiceName, p.PriceRub, start, end, isEndProvided)
	s, err := scanSubscription(row)
	if err != nil {
		if isNoRows(err) {
			return Subscription{}, false, nil
		}
		return Subscription{}, false, err
	}
	return s, true, nil
}

type TotalQueryParams struct {
	UserID      *string
	ServiceName *string
	From        Month
	To          Month
}

func (r *Repo) FindOverlappingForTotal(ctx context.Context, p TotalQueryParams) ([]Subscription, error) {
	const q = `
SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
FROM subscriptions
WHERE ($1::uuid IS NULL OR user_id = $1::uuid)
  AND ($2::text IS NULL OR service_name = $2::text)
  AND start_date <= $4::date
  AND (end_date IS NULL OR end_date >= $3::date)`

	from := p.From.Time()
	to := p.To.Time()

	rows, err := r.pool.Query(ctx, q, p.UserID, p.ServiceName, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Subscription
	for rows.Next() {
		s, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanSubscription(row rowScanner) (Subscription, error) {
	var (
		id          string
		serviceName string
		price       int
		userID      string
		startDate   time.Time
		endDate     *time.Time
		createdAt   time.Time
		updatedAt   time.Time
	)

	if err := row.Scan(&id, &serviceName, &price, &userID, &startDate, &endDate, &createdAt, &updatedAt); err != nil {
		return Subscription{}, err
	}

	startMonth := Month{t: time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)}
	var endMonth *Month
	if endDate != nil {
		m := Month{t: time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)}
		endMonth = &m
	}

	return Subscription{
		ID:          id,
		ServiceName: serviceName,
		PriceRub:    price,
		UserID:      userID,
		StartMonth:  startMonth,
		EndMonth:    endMonth,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

