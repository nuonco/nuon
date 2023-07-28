package jobs

import (
	"context"
	"fmt"

	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
)

func (m *manager) CreateInstall(ctx context.Context, installID string) error {
	m.Opts.ID = installID
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, "installs", m.Opts, "CreateInstall", &jobsv1.CreateInstallRequest{
		InstallId: installID,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger install job: %w", err)
	}

	return nil
}

func (m *manager) DeleteInstall(ctx context.Context, installID string) error {
	m.Opts.ID = installID
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, "installs", m.Opts, "DeleteInstall", &jobsv1.DeleteInstallRequest{
		InstallId: installID,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger install delete job: %w", err)
	}

	return nil
}
