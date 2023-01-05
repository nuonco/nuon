package orgs

import (
	"context"

	"github.com/bufbuild/connect-go"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
)

func (s *server) GetOrg(
	ctx context.Context,
	req *connect.Request[orgv1.GetOrgRequest],
) (*connect.Response[orgv1.GetOrgResponse], error) {
	res := connect.NewResponse(&orgv1.GetOrgResponse{})
	return res, nil
}

func (s *server) GetOrgsByMember(
	ctx context.Context,
	req *connect.Request[orgv1.GetOrgsByMemberRequest],
) (*connect.Response[orgv1.GetOrgsByMemberResponse], error) {
	res := connect.NewResponse(&orgv1.GetOrgsByMemberResponse{})
	return res, nil
}
