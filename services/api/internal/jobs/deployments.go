package jobs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/api/internal/jobs/createdeployment"
)

func (m *manager) CreateDeployment(ctx context.Context, deploymentID string) error {
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, m.Namespace, m.Opts, "CreateDeployment", createdeployment.CreateDeploymentRequest{
		DeploymentID: deploymentID,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger deployment job: %w", err)
	}

	return nil
}
