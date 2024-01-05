package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

func (h *Helpers) ValidateGraph(ctx context.Context, appID string) error {
	_, _, err := h.GetGraph(ctx, appID)
	if err == nil {
		return nil
	}

	edgeErr := &errComponentEdge{}
	if !errors.As(err, &edgeErr) {
		return err
	}

	if errors.Is(err, graph.ErrEdgeAlreadyExists) {
		return stderr.ErrUser{
			Err:         err,
			Description: fmt.Sprintf("dependency between %s and %s already exists", edgeErr.compID, edgeErr.dependencyID),
		}
	}
	if errors.Is(err, graph.ErrEdgeCreatesCycle) {
		return stderr.ErrUser{
			Err:         err,
			Description: fmt.Sprintf("dependency between %s and %s creates a cycle", edgeErr.compID, edgeErr.dependencyID),
		}
	}

	return err
}
