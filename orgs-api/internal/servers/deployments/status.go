package deployments

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	deploymentsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/deployments/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

func (s *server) GetStatus(
	ctx context.Context,
	req *connect.Request[deploymentsv1.GetStatusRequest],
) (*connect.Response[deploymentsv1.GetStatusResponse], error) {
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
func (s *server) getStatus(ctx context.Context, req *deploymentsv1.GetStatusRequest, wkflows workflows.Repo) (*deploymentsv1.GetStatusResponse, error) {
	var status deploymentsv1.Status

	resp, err := wkflows.GetDeploymentProvisionResponse(ctx, req.OrgId, req.AppId, req.ComponentId, req.DeploymentId)
	if err != nil {
		return &deploymentsv1.GetStatusResponse{
			Status: deploymentsv1.Status_STATUS_PROVISIONING,
		}, nil
	}

	switch resp.Status {
	case sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR:
		status = deploymentsv1.Status_STATUS_ERROR
	case sharedv1.ResponseStatus_RESPONSE_STATUS_UNSPECIFIED:
		status = deploymentsv1.Status_STATUS_UNKNOWN
	case sharedv1.ResponseStatus_RESPONSE_STATUS_OK:
		prResp := resp.Response.GetDeploymentStart()
		if prResp == nil {
			status = deploymentsv1.Status_STATUS_UNKNOWN
		} else {
			status = deploymentsv1.Status_STATUS_ACTIVE
		}
	}

	return &deploymentsv1.GetStatusResponse{
		Status: status,
	}, nil
}
