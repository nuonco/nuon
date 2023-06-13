package jobs

import (
	"context"
	"fmt"

	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
)

func (m *manager) CreateDeployment(ctx context.Context, deploymentID string) error {
	m.Opts.ID = deploymentID
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, "deployments", m.Opts, "CreateDeployment", &jobsv1.CreateDeploymentRequest{
		DeploymentId: deploymentID,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger deployment job: %w", err)
	}

	return nil
}
