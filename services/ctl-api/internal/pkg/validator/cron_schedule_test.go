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
		{"every 10m", "10,20,30,40,50 * * * *", false},
		{"exactly 5m apart twice/hour", "0,5 * * * *", false},
		{"5m step with offset", "2-59/5 * * * *", false},
		{"once a day late", "59 23 * * *", false},
		{"crosses day boundary safely", "0 0,23 * * *", false},
		{"business hours every 5m weekdays", "*/5 9-17 * * 1-5", false},
		{"MWF noon only", "0 12 * * 1,3,5", false},
		{"once a year (leap day)", "0 0 29 2 *", false},
		{"once a year (new year)", "0 0 1 1 *", false},

		// fail: too frequent
		{"every minute", "* * * * *", true},
		{"every 1 min explicit", "*/1 * * * *", true},
		{"every 2 min", "*/2 * * * *", true},
		{"every 4 min", "*/4 * * * *", true},
		{"two fires 1 min apart", "0,1 * * * *", true},
		{"three fires 2 min apart", "0,2,4 * * * *", true},
		{"mostly 15m with one 1m gap", "0,15,30,45,46 * * * *", true},
		{"uneven list with 4m gap mid-cycle", "3,13,23,37,41,53 * * * *", true},
		{"4m apart, just under the limit", "0,4 * * * *", true},
		{"hour-boundary 1m gap (59→00)", "0,59 * * * *", true},
		{"step 7 wraps at hour (56→00 = 4m)", "*/7 * * * *", true},
		{"step 3 minute (3m gap)", "*/3 * * * *", true},
		{"two fires close, crossing hour", "55,59 * * * *", true},
		{"frequent only during business hours", "*/3 9-17 * * *", true},

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
