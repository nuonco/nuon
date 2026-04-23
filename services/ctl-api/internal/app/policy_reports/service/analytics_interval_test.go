package service

import (
	"testing"
	"time"
)

func TestIntervalForRange(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		duration  time.Duration
		wantLabel string
		wantExpr  string
	}{
		{"1 hour", time.Hour, "15m", "toStartOfInterval(evaluated_at, INTERVAL 15 MINUTE)"},
		{"3 hours", 3 * time.Hour, "15m", "toStartOfInterval(evaluated_at, INTERVAL 15 MINUTE)"},
		{"4 hours", 4 * time.Hour, "30m", "toStartOfInterval(evaluated_at, INTERVAL 30 MINUTE)"},
		{"6 hours", 6 * time.Hour, "30m", "toStartOfInterval(evaluated_at, INTERVAL 30 MINUTE)"},
		{"exactly 24h", 24 * time.Hour, "hour", "toStartOfHour(evaluated_at)"},
		{"25 hours", 25 * time.Hour, "6h", "toStartOfInterval(evaluated_at, INTERVAL 6 HOUR)"},
		{"exactly 7 days", 7 * 24 * time.Hour, "6h", "toStartOfInterval(evaluated_at, INTERVAL 6 HOUR)"},
		{"8 days", 8 * 24 * time.Hour, "day", "toStartOfDay(evaluated_at)"},
		{"exactly 30 days", 30 * 24 * time.Hour, "day", "toStartOfDay(evaluated_at)"},
		{"31 days", 31 * 24 * time.Hour, "week", "toStartOfWeek(evaluated_at)"},
		{"exactly 90 days", 90 * 24 * time.Hour, "week", "toStartOfWeek(evaluated_at)"},
		{"91 days", 91 * 24 * time.Hour, "month", "toStartOfMonth(evaluated_at)"},
		{"365 days", 365 * 24 * time.Hour, "month", "toStartOfMonth(evaluated_at)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := now
			end := now.Add(tt.duration)
			got := intervalForRange(start, end)

			if got.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", got.Label, tt.wantLabel)
			}
			if expr := got.BucketExpr("evaluated_at"); expr != tt.wantExpr {
				t.Errorf("BucketExpr = %q, want %q", expr, tt.wantExpr)
			}
		})
	}
}
