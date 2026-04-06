package helpers

import (
	"context"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/pkg/errors"
)

// GetConfigSandboxDependentComponentOrder returns the topologically sorted list of component IDs
// that depend on sandbox outputs (directly or transitively through component dependencies).
func (h *Helpers) GetConfigSandboxDependentComponentOrder(ctx context.Context, cfg *app.AppConfig) ([]string, error) {
	sandboxDependent := sandboxDependentComponents(cfg.ComponentConfigConnections)

	// Get the full deploy order, then filter to only sandbox-dependent components
	allOrder, err := h.GetConfigDefaultComponentOrder(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get default component order")
	}

	filtered := make([]string, 0, len(sandboxDependent))
	for _, compID := range allOrder {
		if sandboxDependent[compID] {
			filtered = append(filtered, compID)
		}
	}

	return filtered, nil
}

// sandboxDependentComponents identifies all component IDs that depend on sandbox outputs,
// either directly (via RefTypeSandbox refs) or transitively through component dependencies.
func sandboxDependentComponents(connections []app.ComponentConfigConnection) map[string]bool {
	compRefs := make(map[string][]refs.Ref)
	compDeps := make(map[string][]string)
	for _, ccc := range connections {
		compRefs[ccc.ComponentID] = ccc.Refs
		compDeps[ccc.ComponentID] = ccc.ComponentDependencyIDs
	}

	// Identify components that directly reference sandbox outputs
	sandboxDependent := make(map[string]bool)
	for compID, crefs := range compRefs {
		for _, ref := range crefs {
			if ref.Type == refs.RefTypeSandbox {
				sandboxDependent[compID] = true
				break
			}
		}
	}

	// Build a reverse dependency map: depID -> []componentIDs that depend on it
	reverseDeps := make(map[string][]string)
	for compID, deps := range compDeps {
		for _, dep := range deps {
			reverseDeps[dep] = append(reverseDeps[dep], compID)
		}
	}

	// Walk forward from sandbox-dependent components to include transitive dependents
	queue := make([]string, 0, len(sandboxDependent))
	for compID := range sandboxDependent {
		queue = append(queue, compID)
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, dependent := range reverseDeps[current] {
			if !sandboxDependent[dependent] {
				sandboxDependent[dependent] = true
				queue = append(queue, dependent)
			}
		}
	}

	return sandboxDependent
}
