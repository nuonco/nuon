package helpers

import (
	"context"
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

func (h *Helpers) ValidateGraph(ctx context.Context, appID string) error {
	latestCfg, err := h.GetAppLatestConfig(ctx, appID)
	if err != nil {
		return errors.Wrap(err, "unable to get latest config")
	}

	appCfg, err := h.GetFullAppConfig(ctx, latestCfg.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	_, err = h.GetConfigGraph(ctx, appCfg)
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
