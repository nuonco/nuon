package jobs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/api/internal/jobs/createapp"
)

func (m *manager) CreateApp(ctx context.Context, appID string) error {
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, m.Namespace, m.Opts, "CreateApp", createapp.CreateAppRequest{
		AppID: appID,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger app job: %w", err)
	}

	return nil
}
