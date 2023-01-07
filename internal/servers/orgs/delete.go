package orgs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	orgv1 "github.com/powertoolsdev/protos/api/generated/types/org/v1"
)

func (s *server) DeleteOrg(
	ctx context.Context,
	req *connect.Request[orgv1.DeleteOrgRequest],
) (*connect.Response[orgv1.DeleteOrgResponse], error) {
	deleted, err := s.Svc.DeleteOrg(ctx, req.Msg.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to delete org: %w", err)
	}

	return connect.NewResponse(&orgv1.DeleteOrgResponse{
		Deleted: deleted,
	}), nil
}
