package errparse

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const (
	ErrorMetadataOutput  = "error_output"
	ErrorMetadataMessage = "message"
)

func ParseRunnerJobResult(success bool, errorMetadata map[string]string, runnerJob *app.RunnerJob) (*compositeerrors.CompositeErrorData, error) {
	if success {
		return nil, nil
	}

	raw := RunnerJobErrorText(errorMetadata)
	if raw == "" {
		return nil, nil
	}

	ce := Parse(&ParseContext{
		Raw:       raw,
		Tool:      ToolForRunnerJob(runnerJob),
		Operation: string(runnerJob.Operation),
		Group:     string(runnerJob.Group),
		Owner:     Owner{Type: runnerJob.OwnerType, ID: runnerJob.OwnerID},
		Meta:      errorMetadata,
	})
	if ce == nil {
		return nil, nil
	}

	return compositeerrors.New(ce, compositeerrors.WithSource(runnerJob.OwnerType, runnerJob.OwnerID))
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
