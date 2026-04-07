package subscriptions

import (
	"fmt"
	"time"
)

// Month is a calendar month anchored to its first day (UTC).
// JSON format used in API is "MM-YYYY" (e.g. "07-2025").
type Month struct {
	t time.Time
}

func ParseMonth(value string) (Month, error) {
	if value == "" {
		return Month{}, fmt.Errorf("month is empty")
	}

	parsed, err := time.ParseInLocation("01-2006", value, time.UTC)
	if err != nil {
		return Month{}, fmt.Errorf("invalid month %q, expected MM-YYYY: %w", value, err)
	}

	return Month{t: time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)}, nil
}

func (m Month) String() string {
	if m.t.IsZero() {
		return ""
	}

	return m.t.Format("01-2006")
}

func (m Month) Time() time.Time {
	return m.t
}

func CountMonthsInclusive(from Month, to Month) int {
	if from.t.IsZero() || to.t.IsZero() {
		return 0
	}

	if from.t.After(to.t) {
		return 0
	}

	return (to.t.Year()-from.t.Year())*12 + int(to.t.Month()-from.t.Month()) + 1
}

func (m Month) IsBefore(other Month) bool {
	return !m.t.IsZero() && !other.t.IsZero() && m.t.Before(other.t)
}

func (m Month) IsAfter(other Month) bool {
	return !m.t.IsZero() && !other.t.IsZero() && m.t.After(other.t)
}

func OverlapMonthsInclusive(
	subStart Month,
	subEnd *Month,
	periodFrom Month,
	periodTo Month,
) int {
	if periodFrom.t.After(periodTo.t) {
		return 0
	}

	if subStart.t.IsZero() || periodFrom.t.IsZero() || periodTo.t.IsZero() {
		return 0
	}

	effectiveEnd := periodTo
	if subEnd != nil && !subEnd.t.IsZero() && subEnd.t.Before(effectiveEnd.t) {
		effectiveEnd = *subEnd
	}

	effectiveStart := periodFrom
	if subStart.t.After(effectiveStart.t) {
		effectiveStart = subStart
	}

	if effectiveStart.t.After(effectiveEnd.t) {
		return 0
	}

	return CountMonthsInclusive(effectiveStart, effectiveEnd)
}

