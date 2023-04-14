package jobs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/api/internal/jobs/createinstall"
)

func (m *manager) CreateInstall(ctx context.Context, installID string) error {
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, m.Namespace, m.Opts, "CreateInstall", createinstall.CreateInstallRequest{
		InstallID: installID,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger install job: %w", err)
	}

	return nil
}
