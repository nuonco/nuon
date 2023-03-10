package orgs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

func (s *server) GetStatus(
	ctx context.Context,
	req *connect.Request[orgsv1.GetStatusRequest],
) (*connect.Response[orgsv1.GetStatusResponse], error) {
	wkflowsRepo, err := s.WorkflowsRepo(ctx, req.Msg.OrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflows repo: %w", err)
	}

	resp, err := s.getStatus(ctx, req.Msg, wkflowsRepo)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	return connect.NewResponse(resp), nil
}

//nolint:all
func (s *server) getStatus(ctx context.Context, req *orgsv1.GetStatusRequest, wkflows workflows.Repo) (*orgsv1.GetStatusResponse, error) {
	var status orgsv1.Status

	resp, err := wkflows.GetOrgProvisionResponse(ctx, req.OrgId)
	if err != nil {
		req, err := wkflows.GetOrgProvisionRequest(ctx, req.OrgId)
		if err != nil || req.Request.GetOrgSignup() == nil {
			return &orgsv1.GetStatusResponse{
				Status: orgsv1.Status_STATUS_UNKNOWN,
			}, nil
		}

		return &orgsv1.GetStatusResponse{
			Status: orgsv1.Status_STATUS_PROVISIONING,
		}, nil
	}

	switch resp.Status {
	case sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR:
		status = orgsv1.Status_STATUS_ERROR
	case sharedv1.ResponseStatus_RESPONSE_STATUS_UNSPECIFIED:
		status = orgsv1.Status_STATUS_UNKNOWN
	case sharedv1.ResponseStatus_RESPONSE_STATUS_OK:
		orgResp := resp.Response.GetOrgSignup()
		if orgResp == nil {
			status = orgsv1.Status_STATUS_UNKNOWN
		} else {
			status = orgsv1.Status_STATUS_ACTIVE
		}
	}

	return &orgsv1.GetStatusResponse{
		Status: status,
	}, nil
}
