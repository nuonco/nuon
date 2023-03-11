package orgs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
	orgv1 "github.com/powertoolsdev/mono/pkg/protos/api/generated/types/org/v1"
)

func (s *server) GetOrg(
	ctx context.Context,
	req *connect.Request[orgv1.GetOrgRequest],
) (*connect.Response[orgv1.GetOrgResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	org, err := s.Svc.GetOrg(ctx, req.Msg.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}

	return connect.NewResponse(&orgv1.GetOrgResponse{
		Org: converters.OrgModelToProto(org),
	}), nil
}

func (s *server) GetOrgsByMember(
	ctx context.Context,
	req *connect.Request[orgv1.GetOrgsByMemberRequest],
) (*connect.Response[orgv1.GetOrgsByMemberResponse], error) {
	orgs, _, err := s.Svc.UserOrgs(ctx, req.Msg.MemberId, &models.ConnectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get user orgs: %w", err)
	}

	return connect.NewResponse(&orgv1.GetOrgsByMemberResponse{
		Orgs: converters.OrgModelsToProtos(orgs),
	}), nil
}
