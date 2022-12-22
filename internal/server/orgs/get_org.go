package orgs

import (
	"context"

	"github.com/powertoolsdev/api/internal/request"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
)

func (s *server) GetOrg(ctx context.Context, req *orgv1.GetOrgRequest) (*orgv1.GetOrgResponse, error) {
	orgID, err := request.ParseID(req.OrgId)
	if err != nil {
		return nil, err
	}

	org, err := s.repo.Get(ctx, orgID)
	if err != nil {
		return nil, err
	}

	orgProto, err := orgModelToProto(org)
	if err != nil {
		return nil, err
	}

	return &orgv1.GetOrgResponse{
		Org: orgProto,
	}, nil
}
