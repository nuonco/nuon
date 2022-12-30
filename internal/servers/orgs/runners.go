package orgs

import (
	"context"

	"github.com/bufbuild/connect-go"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func (s *server) GetRunners(
	ctx context.Context,
	req *connect.Request[orgsv1.GetRunnersRequest],
) (*connect.Response[orgsv1.GetRunnersResponse], error) {
	res := connect.NewResponse(&orgsv1.GetRunnersResponse{
		//
	})
	return res, nil
}
