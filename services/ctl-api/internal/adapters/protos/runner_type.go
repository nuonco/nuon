package protos

import (
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func ToRunnerType(runnerType app.AppRunnerType) installsv1.RunnerType {
	switch runnerType {
	case app.AppRunnerTypeAWSECS:
		return installsv1.RunnerType_RUNNER_TYPE_AWS_ECS
	case app.AppRunnerTypeAWSEKS:
		return installsv1.RunnerType_RUNNER_TYPE_AWS_EKS
	case app.AppRunnerTypeAzureAKS:
		return installsv1.RunnerType_RUNNER_TYPE_AZURE_AKS
	case app.AppRunnerTypeAzureACS:
		return installsv1.RunnerType_RUNNER_TYPE_AZURE_ACS
	}

	return installsv1.RunnerType_RUNNER_TYPE_UNSPECIFIED
}
