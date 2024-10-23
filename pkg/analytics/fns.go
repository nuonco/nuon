package analytics

import (
	"context"
	"fmt"

	segment "github.com/segmentio/analytics-go/v3"
)

type (
	GroupFn    func(context.Context) (*segment.Group, error)
	IdentifyFn func(context.Context) (*segment.Identify, error)
	UserIDFn   func(context.Context) (string, error)
)

func NoopIdentifyFn(context.Context) (*segment.Identify, error) {
	return nil, fmt.Errorf("noop")
}

func NoopGroupFn(context.Context) (*segment.Group, error) {
	return nil, fmt.Errorf("noop")
}
