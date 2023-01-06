package orgs

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
)

func (s *server) GetOrg(
	ctx context.Context,
	req *connect.Request[orgv1.GetOrgRequest],
) (*connect.Response[orgv1.GetOrgResponse], error) {
	org, err := s.Svc.GetOrg(ctx, req.Msg.OrgId)
	if err != nil {
		return nil, err
	}

	orgProto, err := servers.OrgModelToProto(org)
	if err != nil {
		return nil, err
	}
	res := connect.NewResponse(&orgv1.GetOrgResponse{
		Org: orgProto,
	})
	return res, nil
}

func (s *server) GetOrgsByMember(
	ctx context.Context,
	req *connect.Request[orgv1.GetOrgsByMemberRequest],
) (*connect.Response[orgv1.GetOrgsByMemberResponse], error) {
	orgs, _, err := s.Svc.UserOrgs(ctx, req.Msg.MemberId, &models.ConnectionOptions{})
	if err != nil {
		return nil, err
	}

	orgProtos := []*orgv1.Org{}

	for _, org := range orgs {
		orgProto, err := servers.OrgModelToProto(org)
		if err != nil {
			return nil, err
		}
		orgProtos = append(orgProtos, orgProto)
	}
	res := connect.NewResponse(&orgv1.GetOrgsByMemberResponse{
		Orgs: orgProtos,
	})
	return res, nil
}
