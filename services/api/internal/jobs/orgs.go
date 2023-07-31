package jobs

import (
	"context"
	"fmt"

	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	enumspb "go.temporal.io/api/enums/v1"
)

func (m *manager) CreateOrg(ctx context.Context, orgID string) error {
	m.Opts.ID = orgID
	m.Opts.WorkflowIDReusePolicy = enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, "orgs", m.Opts, "CreateOrg", &jobsv1.CreateOrgRequest{
		OrgId: orgID,
	})
	if err != nil {
		return fmt.Errorf("unable to provision org: %w", err)
	}

	return nil
}

func (m *manager) DeleteOrg(ctx context.Context, orgID string) error {
	m.Opts.ID = orgID
	m.Opts.WorkflowIDReusePolicy = enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, "orgs", m.Opts, "DeleteOrg", &jobsv1.DeleteOrgRequest{
		OrgId: orgID,
	})
	if err != nil {
		return fmt.Errorf("unable to provision org: %w", err)
	}

	return nil
}
