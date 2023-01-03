package orgs

import (
	"context"

	"github.com/powertoolsdev/api/internal/request"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
)

func (s *server) DeleteOrg(ctx context.Context, req *orgv1.DeleteOrgRequest) (*orgv1.DeleteOrgResponse, error) {
	orgID, err := request.ParseID(req.OrgId)
	if err != nil {
		return &orgv1.DeleteOrgResponse{
			Deleted: false,
		}, err
	}

	deleted, err := s.repo.Delete(ctx, orgID)
	if err != nil {
		return &orgv1.DeleteOrgResponse{
			Deleted: false,
		}, err
	}
	if !deleted {
		return &orgv1.DeleteOrgResponse{
			Deleted: false,
		}, nil
	}

	/*if err := o.wkflowMgr.Deprovision(ctx, orgID.String()); err != nil {
		return false, fmt.Errorf("unable to start deprovision workflow: %w", err)
	}*/
	return &orgv1.DeleteOrgResponse{
		Deleted: true,
	}, nil
}
