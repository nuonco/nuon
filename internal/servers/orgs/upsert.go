package orgs

import (
	"context"

	"github.com/bufbuild/connect-go"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
)

func (s *server) UpsertOrg(
	ctx context.Context,
	req *connect.Request[orgv1.UpsertOrgRequest],
) (*connect.Response[orgv1.UpsertOrgResponse], error) {
	res := connect.NewResponse(&orgv1.UpsertOrgResponse{})
	return res, nil
}
