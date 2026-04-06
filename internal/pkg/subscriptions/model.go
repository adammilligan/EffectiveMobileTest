package subscriptions

import "time"

type Subscription struct {
	ID          string
	ServiceName string
	PriceRub    int
	UserID      string
	StartMonth  Month
	EndMonth    *Month
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

