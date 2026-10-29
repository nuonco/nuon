package config

import (
	"fmt"
	"time"
)

const (
	MinActionTimeout = 30 * time.Minute
	MaxActionTimeout = 24 * time.Hour
)

func ParseActionTimeout(raw string) (time.Duration, error) {
	timeout, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("action_timeout must be a valid duration (for example 30m or 4h): %w", err)
	}
	if timeout < MinActionTimeout || timeout > MaxActionTimeout {
		return 0, fmt.Errorf("action_timeout must be between %s and %s", MinActionTimeout, MaxActionTimeout)
	}
	return timeout, nil
}

func FormatActionTimeout(timeout time.Duration) string {
	if timeout == 0 {
		return ""
	}
	return timeout.String()
}

func ActionTimeoutsEqual(a, b string) bool {
	if a == b {
		return true
	}
	aTimeout, aErr := time.ParseDuration(a)
	bTimeout, bErr := time.ParseDuration(b)
	return aErr == nil && bErr == nil && aTimeout == bTimeout
}
