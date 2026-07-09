package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func TestExitCodeForErr(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "plain error exits 1",
			err:  errors.New("boom"),
			want: 1,
		},
		{
			name: "exit code error controls its code",
			err:  &ui.ErrExitCode{Err: errors.New("builds failed"), Code: "builds_failed", Exit: 3},
			want: 3,
		},
		{
			name: "wrapped exit code error is unwrapped",
			err:  fmt.Errorf("context: %w", &ui.ErrExitCode{Err: errors.New("builds failed"), Code: "builds_failed", Exit: 3}),
			want: 3,
		},
		{
			name: "zero exit code falls back to 1",
			err:  &ui.ErrExitCode{Err: errors.New("boom")},
			want: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := exitCodeForErr(tc.err); got != tc.want {
				t.Fatalf("exitCodeForErr() = %d, want %d", got, tc.want)
			}
		})
	}
}
