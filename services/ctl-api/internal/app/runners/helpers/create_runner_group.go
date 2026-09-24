package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sagikazarmark/slog-shim"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	defaultRunnerGroupHeartBeatTimeout       time.Duration = time.Second * 5
	defaultRunnerGroupSettingsRefreshTimeout time.Duration = time.Minute * 5
)

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (h *Helpers) runnerImageURLForPlatform(platform app.CloudPlatform) string {
	switch platform {
	case app.CloudPlatformGCP:
		if h.cfg.RunnerContainerImageURLGCP != "" {
			return h.cfg.RunnerContainerImageURLGCP
		}
	case app.CloudPlatformAzure:
		if h.cfg.RunnerContainerImageURLAzure != "" {
			return h.cfg.RunnerContainerImageURLAzure
		}
	}
	return h.cfg.RunnerContainerImageURL
}

func (h *Helpers) CreateInstallRunnerGroup(ctx context.Context, install *app.Install) (*app.RunnerGroup, error) {
	ctx = cctx.SetOrgIDContext(ctx, install.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, install.CreatedByID)

	platform := install.AppRunnerConfig.Type
	if install.Org.OrgType != app.OrgTypeDefault || h.cfg.UseLocalRunners {
		platform = app.AppRunnerTypeLocal
	}

	// Install-level sandbox mode takes precedence when set, else fall back to org.
	sandboxMode := install.Org.SandboxMode
	if install.SandboxMode.Valid {
		sandboxMode = install.SandboxMode.Bool
	}

	instanceType := install.AppRunnerConfig.InstanceType
	if instanceType == "" {
		instanceType = app.DefaultInstanceTypeForPlatform(install.AppRunnerConfig.CloudPlatform)
	}

	groups := append(app.CommonRunnerGroupSettingsGroups[:], app.DefaultInstallRunnerGroupSettingsGroups[:]...)
	runnerGroup := app.RunnerGroup{
		OwnerID:   install.ID,
		OwnerType: "installs",
		// OwnerName: install.Name,
		Type:     app.RunnerGroupTypeInstall,
		Platform: install.AppRunnerConfig.Type,
		Runners: []app.Runner{
			{
				Name:              "default",
				DisplayName:       "Default runner",
				Status:            app.RunnerStatusPending,
				StatusDescription: string(app.RunnerStatusPending),
			},
		},
		Settings: app.RunnerGroupSettings{
			SandboxMode:       sandboxMode,
			ContainerImageURL: h.runnerImageURLForPlatform(install.AppRunnerConfig.CloudPlatform),
			ContainerImageTag: h.cfg.RunnerContainerImageTag,
			RunnerAPIURL:      firstNonEmpty(install.AppRunnerConfig.RunnerAPIURL, h.cfg.RunnerAPIURL),
			HeartBeatTimeout:  defaultRunnerGroupHeartBeatTimeout,
			EnableLogging:     true,
			LoggingLevel:      slog.LevelInfo.String(),
			// NOTE(jm): until we add support for writing metrics via our API, this must be disabled as we
			// do not guarantee datadog is running in install accounts.
			EnableMetrics:   false,
			EnableSentry:    true,
			Groups:          groups,
			AWSInstanceType: instanceType,
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

	if err := h.EnsureRunnerSignalsQueue(ctx, runnerGroup.Runners[0].ID); err != nil {
		return nil, fmt.Errorf("unable to create runner signals queue: %w", err)
	}

	if err := h.CreateRunnerQueues(ctx, &runnerGroup.Runners[0], &runnerGroup.Settings); err != nil {
		return nil, fmt.Errorf("unable to create runner queues: %w", err)
	}

	return &runnerGroup, nil
}
