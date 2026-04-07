package subscriptions

import "testing"

func TestParseMonth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		isWantErr bool
		want      string
	}{
		{name: "ok", input: "07-2025", want: "07-2025"},
		{name: "empty", input: "", isWantErr: true},
		{name: "bad format", input: "2025-07", isWantErr: true},
		{name: "bad month", input: "13-2025", isWantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m, err := ParseMonth(tt.input)
			if tt.isWantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := m.String(); got != tt.want {
				t.Fatalf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func TestOverlapMonthsInclusive(t *testing.T) {
	t.Parallel()

	parse := func(t *testing.T, v string) Month {
		t.Helper()

		m, err := ParseMonth(v)
		if err != nil {
			t.Fatalf("parse month: %v", err)
		}

		return m
	}

	from := parse(t, "07-2025")
	to := parse(t, "09-2025")

	t.Run("subscription inside period", func(t *testing.T) {
		t.Parallel()
		start := parse(t, "08-2025")
		end := parse(t, "08-2025")

		got := OverlapMonthsInclusive(start, &end, from, to)
		if got != 1 {
			t.Fatalf("want 1, got %d", got)
		}
	})

	t.Run("open ended overlaps", func(t *testing.T) {
		t.Parallel()
		start := parse(t, "08-2025")

		got := OverlapMonthsInclusive(start, nil, from, to)
		if got != 2 {
			t.Fatalf("want 2, got %d", got)
		}
	})

	t.Run("no overlap", func(t *testing.T) {
		t.Parallel()
		start := parse(t, "01-2020")
		end := parse(t, "02-2020")

		got := OverlapMonthsInclusive(start, &end, from, to)
		if got != 0 {
			t.Fatalf("want 0, got %d", got)
		}
	})
}

