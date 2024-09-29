package protos

import (
	"fmt"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Adapter) toRunnerSettings(runner *app.Runner, apiToken string) *planv1.RunnerSettings {
	return &planv1.RunnerSettings{
		ApiToken: apiToken,
		ApiUrl:   runner.RunnerGroup.Settings.RunnerAPIURL,
		Image: &planv1.RunnerImage{
			Tag: runner.RunnerGroup.Settings.ContainerImageTag,
			Url: runner.RunnerGroup.Settings.ContainerImageURL,
		},
	}
}

func (a *Adapter) toRunnerType(runner *app.Runner) planv1.RunnerType {
	switch runner.RunnerGroup.Platform {
	case app.AppRunnerTypeAWSECS:
		return planv1.RunnerType_RUNNER_TYPE_AWS_ECS
	case app.AppRunnerTypeAWSEKS:
		return planv1.RunnerType_RUNNER_TYPE_AWS_EKS
	case app.AppRunnerTypeAzureAKS:
		return planv1.RunnerType_RUNNER_TYPE_AZURE_AKS
	case app.AppRunnerTypeAzureACS:
		return planv1.RunnerType_RUNNER_TYPE_AZURE_ACS
	}

	return planv1.RunnerType_RUNNER_TYPE_UNSPECIFIED
}

func (a *Adapter) ToRunnerInstallPlanRequest(runner *app.Runner, install *app.Install, apiToken string) (*planv1.CreatePlanRequest, error) {
	sandboxSettings, err := a.toSandboxSettings(install)
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox settings: %w", err)
	}

	return &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Runner{
			Runner: &planv1.RunnerInput{
				OrgId:           runner.OrgID,
				AppId:           install.AppID,
				InstallId:       install.ID,
				RunnerId:        runner.ID,
				Type:            a.toRunnerType(runner),
				SandboxSettings: sandboxSettings,
				RunnerSettings:  a.toRunnerSettings(runner, apiToken),
				AwsSettings:     a.toAWSSettings(install),
				AzureSettings:   a.toAzureSettings(install),
			},
		},
	}, nil
}
