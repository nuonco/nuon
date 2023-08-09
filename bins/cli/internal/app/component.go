package app

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/gqlclient"
	"github.com/powertoolsdev/mono/pkg/ui"
)

func (c *commands) getComponent(ctx context.Context, componentID, componentName string) (*gqlclient.Component, error) {
	if componentID != "" {
		compResp, err := c.apiClient.GetComponent(ctx, componentID)
		if err != nil {
			return nil, fmt.Errorf("unable to get component: %w", err)
		}

		return compResp, nil
	}

	ui.Step(ctx, "fetching component id for component %s", componentName)
	compsResp, err := c.apiClient.GetComponents(ctx, c.appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get components: %w", err)
	}

	for _, comp := range compsResp {
		if comp.Name == componentName {
			break
		}
		componentID = comp.Id
	}
	if componentID == "" {
		return nil, fmt.Errorf("unable to map component name to component id")
	}
	compResp, err := c.apiClient.GetComponent(ctx, componentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	return compResp, nil
}
