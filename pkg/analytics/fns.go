package analytics

import (
	"context"

	segment "github.com/segmentio/analytics-go/v3"
)

type (
	GroupFn    func(context.Context) (*segment.Group, error)
	IdentifyFn func(context.Context) (*segment.Identify, error)
	UserIDFn   func(context.Context) (string, error)
)
