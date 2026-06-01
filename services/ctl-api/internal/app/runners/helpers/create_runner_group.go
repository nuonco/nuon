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
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const (
	defaultRunnerGroupHeartBeatTimeout       time.Duration = time.Second * 5
	defaultRunnerGroupSettingsRefreshTimeout time.Duration = time.Minute * 5
)

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

	parallelJobs, err := h.featuresClient.OrgHasFeature(ctx, install.OrgID, app.OrgFeatureParallelRunnerJobs)
	if err != nil {
		return nil, fmt.Errorf("unable to check parallel runner jobs feature: %w", err)
	}
	if parallelJobs {
		if err := h.CreateRunnerQueues(ctx, &runnerGroup.Runners[0], &runnerGroup.Settings); err != nil {
			return nil, fmt.Errorf("unable to create runner queues: %w", err)
		}
	}

	// Legacy evClient.Send removed — event loop system has been removed.

	return &runnerGroup, nil
}

func (h *Helpers) CreateOrgRunnerGroup(ctx context.Context, org *app.Org) (*app.RunnerGroup, error) {
	ctx = cctx.SetOrgIDContext(ctx, org.ID)
	ctx = cctx.SetAccountIDContext(ctx, org.CreatedByID)

	platform := app.AppRunnerTypeAWSEKS
	controlPlaneCloud := app.CloudPlatformAWS
	if h.cfg.IsGCP() {
		platform = app.AppRunnerTypeGCPGKE
		controlPlaneCloud = app.CloudPlatformGCP
	}
	if h.cfg.IsAzure() {
		platform = app.AppRunnerTypeAzureAKS
		controlPlaneCloud = app.CloudPlatformAzure
	}
	if org.OrgType != app.OrgTypeDefault || h.cfg.UseLocalRunners {
		platform = app.AppRunnerTypeLocal
	}

	// Build cloud-specific identity for the org runner service account
	var orgAWSIAMRoleARN string
	var orgGCPServiceAccount string
	var orgAzureClientID string
	switch h.cfg.CloudProvider {
	case string(app.CloudPlatformGCP):
		orgGCPServiceAccount = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", org.ID, h.cfg.ManagementAccountID)
	case string(app.CloudPlatformAzure):
		// Azure per-org managed identity is created by ProvisionIAM workflow.
		// OrgAzureClientID is populated after IAM provisioning completes.
	default:
		orgAWSIAMRoleARN = fmt.Sprintf("arn:aws:iam::%s:role/orgs/%s/runner-%s", h.cfg.ManagementAccountID, org.ID, org.ID)
	}

	groups := append(app.CommonRunnerGroupSettingsGroups[:], app.DefaultOrgRunnerGroupSettingsGroups[:]...)
	runnerGroup := app.RunnerGroup{
		OwnerID:   org.ID,
		OwnerType: "orgs",
		// OwnerName: org.Name,
		Type:     app.RunnerGroupTypeOrg,
		Platform: platform,
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
			ContainerImageURL: h.runnerImageURLForPlatform(controlPlaneCloud),
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
			}),

			OrgAWSIAMRoleARN:         orgAWSIAMRoleARN,
			OrgGCPServiceAccount:     orgGCPServiceAccount,
			OrgAzureClientID:         orgAzureClientID,
			LocalAWSIAMRoleARN:       "",
			OrgK8sServiceAccountName: fmt.Sprintf("runner-%s", org.ID),
		},
	}
	res := h.db.WithContext(ctx).Create(&runnerGroup)
	if res.Error != nil {
		return nil, res.Error
	}

	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     runnerGroup.Runners[0].ID,
		OwnerType:   "runners",
		Namespace:   "runners",
		Name:        "runner-signals",
		MaxInFlight: 10,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create runner queue: %w", err)
	}

	if err := h.CreateRunnerQueues(ctx, &runnerGroup.Runners[0], &runnerGroup.Settings); err != nil {
		return nil, fmt.Errorf("unable to create runner queues: %w", err)
	}

	return &runnerGroup, nil
}
