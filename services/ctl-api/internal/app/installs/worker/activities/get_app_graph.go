package activities

import (
	"context"
	"fmt"

	"github.com/dominikbraun/graph"
)

type GetAppGraphRequest struct {
	AppID    string `json:"app_id"`
	Inverted bool   `json:"inverted"`
}

func (a *Activities) GetAppGraph(ctx context.Context, req GetAppGraphRequest) ([]string, error) {
	g, rootIDs, err := a.appsHelpers.GetGraph(ctx, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("unable to get graph: %w", err)
	}

	componentIDs := make([]string, 0)
	for _, rootID := range rootIDs {
		if err := graph.BFS(g, rootID, func(compID string) bool {
			componentIDs = append(componentIDs, compID)
			return false
		}); err != nil {
			return nil, fmt.Errorf("unable to build app graph: %w", err)
		}
	}

	return componentIDs, nil
}
