package jobs

import (
	"context"
	"fmt"

	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	enumspb "go.temporal.io/api/enums/v1"
)

func (m *manager) CreateApp(ctx context.Context, appID string) error {
	m.Opts.ID = appID
	m.Opts.WorkflowIDReusePolicy = enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING

	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, "apps", m.Opts, "CreateApp", &jobsv1.CreateAppRequest{
		AppId: appID,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger app job: %w", err)
	}

	return nil
}
