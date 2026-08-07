package errparse

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const (
	ErrorMetadataOutput  = "error_output"
	ErrorMetadataMessage = "message"
)

func ParseRunnerJobResult(success bool, errorMetadata map[string]string, runnerJob *app.RunnerJob, resolveProvider func() Provider) (*compositeerrors.CompositeErrorData, error) {
	if success {
		return nil, nil
	}

	raw := RunnerJobErrorText(errorMetadata)
	if raw == "" {
		return nil, nil
	}

	ce := Parse(&ParseContext{
		Raw:             raw,
		Tool:            ToolForRunnerJob(runnerJob),
		Operation:       string(runnerJob.Operation),
		Group:           string(runnerJob.Group),
		Owner:           Owner{Type: runnerJob.OwnerType, ID: runnerJob.OwnerID},
		Meta:            errorMetadata,
		ResolveProvider: resolveProvider,
	})
	if ce == nil {
		return nil, nil
	}

	return compositeerrors.New(ce, compositeerrors.WithSource(runnerJob.OwnerType, runnerJob.OwnerID))
}

func ResolveRunnerJobProvider(ctx context.Context, db *gorm.DB, runnerJob *app.RunnerJob) Provider {
	if db == nil || runnerJob == nil || runnerJob.RunnerID == "" {
		return ProviderUnknown
	}

	var runner app.Runner
	if err := db.WithContext(ctx).
		Select("id", "runner_group_id").
		Preload("RunnerGroup", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "platform")
		}).
		Where(&app.Runner{ID: runnerJob.RunnerID}).
		Take(&runner).Error; err != nil {
		return ProviderUnknown
	}

	switch runner.RunnerGroup.Platform.CloudPlatform() {
	case app.CloudPlatformAWS:
		return ProviderAWS
	case app.CloudPlatformAzure:
		return ProviderAzure
	case app.CloudPlatformGCP:
		return ProviderGCP
	default:
		return ProviderUnknown
	}
}

func RunnerJobErrorText(errorMetadata map[string]string) string {
	for _, key := range []string{ErrorMetadataOutput, ErrorMetadataMessage} {
		if value := errorMetadata[key]; value != "" {
			return value
		}
	}
	return ""
}

func ToolForRunnerJob(runnerJob *app.RunnerJob) Tool {
	switch runnerJob.Type {
	case app.RunnerJobTypeTerraformDeploy,
		app.RunnerJobTypeTerraformModuleBuild,
		app.RunnerJobTypeSandboxTerraform,
		app.RunnerJobTypeSandboxTerraformPlan,
		app.RunnerJobTypeRunnerTerraform:
		return ToolTerraform
	case app.RunnerJobTypeHelmChartDeploy,
		app.RunnerJobTypeHelmChartBuild,
		app.RunnerJobTypeRunnerHelm:
		return ToolHelm
	case app.RunnerJobTypePulumiDeploy,
		app.RunnerJobTypePulumiBuild,
		app.RunnerJobTypeSandboxPulumi:
		return ToolPulumi
	case app.RunnerJobTypeDockerBuild:
		return ToolDocker
	case app.RunnerJobTypeKubrenetesManifestDeploy,
		app.RunnerJobTypeKubernetesManifestBuild:
		return ToolKubernetes
	case app.RunnerJobTypeContainerImageBuild,
		app.RunnerJobTypeOCISync:
		return ToolOCI
	default:
		return ToolUnknown
	}
}
