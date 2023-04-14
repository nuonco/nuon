package jobs

import (
	"context"
	"fmt"

	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
)

func (m *manager) CreateInstall(ctx context.Context, installID string) error {
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, m.Namespace, m.Opts, "CreateInstall", &jobsv1.CreateInstallRequest{
		InstallId: installID,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger install job: %w", err)
	}

	return nil
}
