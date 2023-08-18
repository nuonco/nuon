package app

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/deprecated/api/gqlclient"
	"github.com/powertoolsdev/mono/pkg/ui"
	"golang.org/x/sync/errgroup"
)

type DeployOpts struct {
	InstallID string
	All       bool

	BuildID string
	Latest  bool

	ComponentName string
	ComponentID   string
}

func (c *commands) Deploy(ctx context.Context, opts *DeployOpts) error {
	if err := c.ensureAppID(); err != nil {
		return err
	}
	if opts.BuildID == "" && !opts.Latest {
		return fmt.Errorf("must set either -latest or a build id")
	}
	if opts.InstallID == "" && !opts.All {
		return fmt.Errorf("must set either -latest or an install id")
	}

	var buildID string
	var componentID string
	if opts.BuildID != "" {
		ui.Step(ctx, "using build id %s", opts.BuildID)
		buildID = opts.BuildID

		buildResp, err := c.apiClient.GetBuild(ctx, buildID)
		if err != nil {
			return fmt.Errorf("unable to get build")
		}
		componentID = buildResp.ComponentId
	} else if opts.Latest {
		ui.Step(ctx, "fetching latest build")
		if opts.ComponentID == "" && opts.ComponentName == "" {
			return fmt.Errorf("either a component ID or name must be passed in when using --latest")
		}

		comp, err := c.getComponent(ctx, opts.ComponentID, opts.ComponentName)
		if err != nil {
			return err
		}
		componentID = comp.Id

		builds, err := c.apiClient.GetBuilds(ctx, comp.Id)
		if err != nil {
			return fmt.Errorf("unable to get builds for component: %w", err)
		}
		buildID = builds[len(builds)-1].Id

		ui.Step(ctx, "using build id %s", buildID)
	}

	// fetch installs
	installIDs := make([]string, 0)
	if opts.InstallID != "" {
		installIDs = append(installIDs, opts.InstallID)
	} else if opts.All {
		ui.Step(ctx, "deploying to all installs")
		installs, err := c.apiClient.GetInstalls(ctx, c.appID)
		if err != nil {
			return fmt.Errorf("unable to get installs: %w", err)
		}

		for _, inst := range installs {
			installIDs = append(installIDs, inst.Id)
		}
	}

	ui.Step(ctx, "deploying to %d installs", len(installIDs))
	grp, ctx := errgroup.WithContext(ctx)

	for _, instID := range installIDs {
		installID := instID
		grp.Go(func() error {
			deployResp, err := c.apiClient.StartDeploy(ctx, gqlclient.DeployInput{
				BuildId:     buildID,
				InstallId:   installID,
				ComponentId: componentID,
			})
			if err != nil {
				return fmt.Errorf("unable to create deploy: %w", err)
			}

			for {
				status, err := c.apiClient.GetInstanceStatus(ctx, installID, componentID, deployResp.Id)
				if err != nil {
					return fmt.Errorf("unable to get deploy status: %w", err)
				}

				if status == gqlclient.StatusActive {
					break
				}
				if status == gqlclient.StatusError {
					return fmt.Errorf("build failed")
				}

				if status == gqlclient.StatusProvisioning {
					ui.Step(ctx, "deploy %s still deploying", deployResp.Id)
				}

				time.Sleep(time.Second * 1)
			}
			return nil
		})
	}

	if err := grp.Wait(); err != nil {
		return fmt.Errorf("not all deploys succeeded: %w", err)
	}
	return nil
}
