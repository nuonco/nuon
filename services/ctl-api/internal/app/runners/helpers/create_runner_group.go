package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sagikazarmark/slog-shim"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

const (
	defaultRunnerGroupHeartBeatTimeout       time.Duration = time.Second * 5
	defaultRunnerGroupSettingsRefreshTimeout time.Duration = time.Minute * 5
)

func (h *Helpers) CreateInstallRunnerGroup(ctx context.Context, install *app.Install) (*app.RunnerGroup, error) {
	ctx = cctx.SetOrgIDContext(ctx, install.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, install.CreatedByID)

	platform := install.AppRunnerConfig.Type
	if install.Org.OrgType != app.OrgTypeDefault || h.cfg.UseLocalRunners {
		platform = app.AppRunnerTypeLocal
	}

	groups := append(app.CommonRunnerGroupSettingsGroups[:], app.DefaultInstallRunnerGroupSettingsGroups[:]...)
	runnerGroup := app.RunnerGroup{
		OwnerID:   install.ID,
		OwnerType: "installs",
		Type:      app.RunnerGroupTypeInstall,
		Platform:  install.AppRunnerConfig.Type,
		Runners: []app.Runner{
			{
				Name:              "default",
				DisplayName:       "Default runner",
				Status:            app.RunnerStatusPending,
				StatusDescription: string(app.RunnerStatusPending),
			},
		},
		Settings: app.RunnerGroupSettings{
			SandboxMode:       install.Org.SandboxMode,
			ContainerImageURL: h.cfg.RunnerContainerImageURL,
			ContainerImageTag: h.cfg.RunnerContainerImageTag,
			RunnerAPIURL:      h.cfg.RunnerAPIURL,
			HeartBeatTimeout:  defaultRunnerGroupHeartBeatTimeout,
			EnableLogging:     true,
			LoggingLevel:      slog.LevelInfo.String(),
			// NOTE(jm): until we add support for writing metrics via our API, this must be disabled as we
			// do not guarantee datadog is running in install accounts.
			EnableMetrics:   false,
			EnableSentry:    true,
			Groups:          groups,
			AWSInstanceType: "t3a.medium",
			Metadata: pgtype.Hstore(map[string]*string{
				"org.id":          generics.ToPtr(install.OrgID),
				"org.name":        generics.ToPtr(install.Org.Name),
				"org.type":        generics.ToPtr(string(install.Org.OrgType)),
				"app.id":          generics.ToPtr(install.AppID),
				"install.id":      generics.ToPtr(install.ID),
				"runner.type":     generics.ToPtr(string(app.RunnerGroupTypeInstall)),
				"runner.platform": generics.ToPtr(string(platform)),
				"env":             generics.ToPtr(string(h.cfg.Env)),
				// NOTE(jm): we also set the runner group at create time
			}),
		},
	}

	res := h.db.WithContext(ctx).Create(&runnerGroup)
	if res.Error != nil {
		return nil, res.Error
	}

	h.evClient.Send(ctx, runnerGroup.Runners[0].ID, &signals.Signal{
		Type: signals.OperationCreated,
	})

	return &runnerGroup, nil
}

func (h *Helpers) CreateOrgRunnerGroup(ctx context.Context, org *app.Org) (*app.RunnerGroup, error) {
	ctx = cctx.SetOrgIDContext(ctx, org.ID)
	ctx = cctx.SetAccountIDContext(ctx, org.CreatedByID)

	platform := app.AppRunnerTypeAWSEKS
	if org.OrgType != app.OrgTypeDefault || h.cfg.UseLocalRunners {
		platform = app.AppRunnerTypeLocal
	}

	groups := append(app.CommonRunnerGroupSettingsGroups[:], app.DefaultOrgRunnerGroupSettingsGroups[:]...)
	runnerGroup := app.RunnerGroup{
		OwnerID:   org.ID,
		OwnerType: "orgs",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  platform,
		Runners: []app.Runner{
			{
				Name:              "default",
				DisplayName:       "Default runner",
				Status:            app.RunnerStatusPending,
				StatusDescription: string(app.RunnerStatusPending),
			},
		},
		Settings: app.RunnerGroupSettings{
			SandboxMode:       org.SandboxMode,
			ContainerImageURL: h.cfg.RunnerContainerImageURL,
			ContainerImageTag: h.cfg.RunnerContainerImageTag,
			RunnerAPIURL:      h.cfg.RunnerAPIURL,
			HeartBeatTimeout:  defaultRunnerGroupHeartBeatTimeout,
			EnableLogging:     true,
			LoggingLevel:      slog.LevelInfo.String(),
			EnableMetrics:     true,
			EnableSentry:      true,
			Groups:            groups,
			Metadata: pgtype.Hstore(map[string]*string{
				"org.id":          generics.ToPtr(org.ID),
				"org.name":        generics.ToPtr(org.Name),
				"org.type":        generics.ToPtr(string(org.OrgType)),
				"runner.type":     generics.ToPtr(string(app.RunnerGroupTypeInstall)),
				"runner.platform": generics.ToPtr(string(platform)),
				"env":             generics.ToPtr(string(h.cfg.Env)),
				// NOTE(jm): we also set the runner group at create time
			}),

			// NOTE(jm): this is mainly a legacy relic, where instead of actually tracking infra resources in our API, via a
			// catalog, we actually pass around templates for IAM role ARNs
			OrgAWSIAMRoleARN:         "",
			LocalAWSIAMRoleARN:       "",
			OrgK8sServiceAccountName: fmt.Sprintf("runner-%s", org.ID),
		},
	}
	res := h.db.WithContext(ctx).Create(&runnerGroup)
	if res.Error != nil {
		return nil, res.Error
	}

	h.evClient.Send(ctx, runnerGroup.Runners[0].ID, &signals.Signal{
		Type: signals.OperationCreated,
	})
	return &runnerGroup, nil
}
