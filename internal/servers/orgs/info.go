package orgs

import (
	"context"

	"github.com/bufbuild/connect-go"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func (s *server) GetInfo(
	ctx context.Context,
	req *connect.Request[orgsv1.GetInfoRequest],
) (*connect.Response[orgsv1.GetInfoResponse], error) {
	ctx, err := s.CtxProvider.SetContext(ctx, req.Msg.OrgId)
	if err != nil {
		return nil, err
	}

	resp, err := s.Svc.GetInfo(ctx, req.Msg.OrgId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}
