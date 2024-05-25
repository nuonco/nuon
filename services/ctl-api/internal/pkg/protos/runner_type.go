package protos

import (
	contextv1 "github.com/powertoolsdev/mono/pkg/types/components/context/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func ToRunnerType(runnerType app.AppRunnerType) contextv1.RunnerType {
	switch runnerType {
	case app.AppRunnerTypeAWSECS:
		return contextv1.RunnerType_RUNNER_TYPE_AWS_ECS
	case app.AppRunnerTypeAWSEKS:
		return contextv1.RunnerType_RUNNER_TYPE_AWS_EKS
	case app.AppRunnerTypeAzureAKS:
		return contextv1.RunnerType_RUNNER_TYPE_AZURE_AKS
	case app.AppRunnerTypeAzureACS:
		return contextv1.RunnerType_RUNNER_TYPE_AZURE_ACS
	}

	return contextv1.RunnerType_RUNNER_TYPE_UNSPECIFIED
}
