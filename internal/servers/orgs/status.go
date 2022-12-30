package orgs

import (
	"context"

	"github.com/bufbuild/connect-go"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func (s *server) GetStatus(
	ctx context.Context,
	req *connect.Request[orgsv1.GetStatusRequest],
) (*connect.Response[orgsv1.GetStatusResponse], error) {
	res := connect.NewResponse(&orgsv1.GetStatusResponse{
		//
	})
	return res, nil
}
