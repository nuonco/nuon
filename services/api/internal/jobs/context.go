package jobs

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/temporal/client"
)

// FromContext returns a new job manager from the provided context
func FromContext(ctx context.Context) (*manager, error) {
	val := ctx.Value(temporal.ContextKey{})
	temporalClient, ok := val.(temporal.Client)
	if !ok {
		return nil, fmt.Errorf("no temporal client configured in context")
	}

	v := validator.New()
	mgr, err := New(v, WithClient(temporalClient))
	if err != nil {
		return nil, fmt.Errorf("unable to get manager: %w", err)
	}

	return mgr, nil
}
