package jobs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/api/internal/jobs/createorg"
)

func (m *manager) CreateOrg(ctx context.Context, orgID string) error {
	_, err := m.Client.ExecuteWorkflowInNamespace(ctx, m.Namespace, m.Opts, "CreateOrg", createorg.CreateOrgRequest{
		OrgID: orgID,
	})
	if err != nil {
		return fmt.Errorf("unable to provision org: %w", err)
	}

	return nil
}
