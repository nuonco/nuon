package app

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/deprecated/api/gqlclient"
	"github.com/powertoolsdev/mono/pkg/ui"
)

type BuildOpts struct {
	GitRef string

	ComponentName string
	ComponentID   string
}

func (c *commands) Build(ctx context.Context, opts *BuildOpts) error {
	if err := c.ensureAppID(); err != nil {
		return err
	}

	componentID := opts.ComponentID
	if componentID == "" {
		ui.Step(ctx, "fetching component id for component %s", opts.ComponentName)
		compResp, err := c.apiClient.GetComponents(ctx, c.appID)
		if err != nil {
			return fmt.Errorf("unable to get components: %w", err)
		}

		for _, comp := range compResp {
			if comp.Name == opts.ComponentName {
				break
			}
			componentID = comp.Id
		}
	}
	if componentID == "" {
		return fmt.Errorf("unable to map component name to component id")
	}
	compResp, err := c.apiClient.GetComponent(ctx, componentID)
	if err != nil {
		return fmt.Errorf("unable to get component: %w", err)
	}

	buildResp, err := c.apiClient.StartBuild(ctx, gqlclient.BuildInput{
		GitRef:      opts.GitRef,
		ComponentId: componentID,
	})
	if err != nil {
		return fmt.Errorf("unable to start build: %w", err)
	}
	ui.Step(ctx, "successfully created build: %s", buildResp.Id)

	for {
		status, err := c.apiClient.GetBuildStatus(ctx, buildResp.Id)
		if err != nil {
			return fmt.Errorf("unable to get build status: %w", err)
		}

		if status == gqlclient.StatusActive {
			break
		}
		if status == gqlclient.StatusError {
			return fmt.Errorf("build failed")
		}

		if status == gqlclient.StatusProvisioning {
			ui.Step(ctx, "build %s still building", buildResp.Id)
		}

		time.Sleep(time.Second * 1)
	}

	ui.Step(ctx, "successfully built component: %s", compResp.Name)
	return nil
}
