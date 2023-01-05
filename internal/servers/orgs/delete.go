package orgs

import (
	"context"

	"github.com/bufbuild/connect-go"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
)

func (s *server) DeleteOrg(
	ctx context.Context,
	req *connect.Request[orgv1.DeleteOrgRequest],
) (*connect.Response[orgv1.DeleteOrgResponse], error) {
	res := connect.NewResponse(&orgv1.DeleteOrgResponse{})
	return res, nil
}
