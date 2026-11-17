package service

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

func (s *service) mcpWorkflowContinuation(ctx context.Context, orgID, workflowID string, operation permissions.PermissionID) error {
	if !authz.HasAuthority(ctx) {
		return nil
	}
	resources, err := access.Workflows.In(s.db, orgID)
	if err != nil {
		return err
	}
	target, err := resources.Resolve(ctx, workflowID)
	if err != nil {
		return err
	}
	requirements, err := s.workflowRequirements(ctx, target, operation)
	if err != nil {
		return err
	}
	for _, requirement := range requirements {
		if err := require.MCPOperation(ctx, requirement.Operation, requirement.Target); err != nil {
			return err
		}
	}
	return nil
}
