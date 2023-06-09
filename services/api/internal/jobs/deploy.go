package jobs

import (
	"context"
	"fmt"

	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
)

func (m *manager) StartDeploy(ctx context.Context, deployID string) (string, error) {
	m.Opts.ID = deployID
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, "deploys", m.Opts, "StartDeploy", &jobsv1.StartDeployRequest{
		DeployId: deployID,
	})
	if err != nil {
		return "", fmt.Errorf("unable to trigger deploy job: %w", err)
	}

	return "deploy_id", nil
}
