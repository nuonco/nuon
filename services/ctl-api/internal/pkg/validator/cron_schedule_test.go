package validator

import (
	"testing"
)

func TestCronSchedule_MinInterval(t *testing.T) {
	v := New()

	cases := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		// pass
		{"empty", "", false},
		{"every 5 min", "*/5 * * * *", false},
		{"hourly on 0", "0 * * * *", false},
		{"every 15 min", "*/15 * * * *", false},
		{"weekdays 9am", "0 9 * * 1-5", false},
		{"sunday midnight", "0 0 * * 0", false},
		{"daily midnight", "0 0 * * *", false},

		// fail: too frequent
		{"every minute", "* * * * *", true},
		{"every 1 min explicit", "*/1 * * * *", true},
		{"every 2 min", "*/2 * * * *", true},
		{"every 4 min", "*/4 * * * *", true},
		{"two fires 1 min apart", "0,1 * * * *", true},
		{"three fires 2 min apart", "0,2,4 * * * *", true},
		{"mostly 15m with one 1m gap", "0,15,30,45,46 * * * *", true},

		// fail: invalid syntax
		{"invalid", "not a cron", true},
		{"too few fields", "* * *", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CronSchedule(v, tc.expr)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q, got nil", tc.expr)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error for %q, got %v", tc.expr, err)
			}
		})
	}
}
