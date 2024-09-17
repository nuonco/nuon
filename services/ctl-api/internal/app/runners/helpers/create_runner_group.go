package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

const (
	defaultRunnerGroupHeartBeatTimeout       time.Duration = time.Second * 5
	defaultRunnerGroupSettingsRefreshTimeout time.Duration = time.Minute * 5
)

func (h *Helpers) CreateInstallRunnerGroup(ctx context.Context, install *app.Install) (*app.RunnerGroup, error) {
	ctx = middlewares.SetOrgIDContext(ctx, install.OrgID)
	ctx = middlewares.SetAccountIDContext(ctx, install.CreatedByID)

	rg, err := h.createRunnerGroup(ctx, install.ID, "installs", app.RunnerGroupTypeInstall)
	if err != nil {
		return nil, fmt.Errorf("unable to create runner group: %w", err)
	}

	h.evClient.Send(ctx, rg.Runners[0].ID, &signals.Signal{
		Type: signals.OperationCreated,
	})

	return rg, nil
}

func (h *Helpers) CreateOrgRunnerGroup(ctx context.Context, org *app.Org) (*app.RunnerGroup, error) {
	ctx = middlewares.SetOrgIDContext(ctx, org.ID)
	ctx = middlewares.SetAccountIDContext(ctx, org.CreatedByID)

	rg, err := h.createRunnerGroup(ctx, org.ID, "orgs", app.RunnerGroupTypeOrg)
	if err != nil {
		return nil, fmt.Errorf("unable to create runner group: %w", err)
	}

	h.evClient.Send(ctx, rg.Runners[0].ID, &signals.Signal{
		Type: signals.OperationCreated,
	})

	return rg, nil
}

func (h *Helpers) createRunnerGroup(ctx context.Context, ownerID, ownerType string, runnerGroupTyp app.RunnerGroupType) (*app.RunnerGroup, error) {
	runnerGroup := app.RunnerGroup{
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Type:      runnerGroupTyp,
		Runners: []app.Runner{
			{
				Name:              "default",
				DisplayName:       "Default runner",
				Status:            app.RunnerStatusPending,
				StatusDescription: string(app.RunnerStatusPending),
			},
		},
		Settings: app.RunnerGroupSettings{
			ContainerImageURL:      h.cfg.RunnerContainerImageURL,
			ContainerImageTag:      h.cfg.RunnerContainerImageTag,
			RunnerAPIURL:           h.cfg.RunnerAPIURL,
			HeartBeatTimeout:       defaultRunnerGroupHeartBeatTimeout,
			SettingsRefreshTimeout: defaultRunnerGroupSettingsRefreshTimeout,
		},
	}
	res := h.db.WithContext(ctx).Create(&runnerGroup)
	if res.Error != nil {
		return nil, res.Error
	}

	return &runnerGroup, nil
}
