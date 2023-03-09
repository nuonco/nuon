package instances

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	instancesv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/instances/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

func (s *server) GetStatus(
	ctx context.Context,
	req *connect.Request[instancesv1.GetStatusRequest],
) (*connect.Response[instancesv1.GetStatusResponse], error) {
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
func (s *server) getStatus(ctx context.Context, req *instancesv1.GetStatusRequest, wkflows workflows.Repo) (*instancesv1.GetStatusResponse, error) {
	var status instancesv1.Status

	resp, err := wkflows.GetInstanceProvisionResponse(ctx, req.OrgId, req.AppId, req.ComponentId, req.DeploymentId, req.InstallId)
	if err != nil {
		return &instancesv1.GetStatusResponse{
			Status: instancesv1.Status_STATUS_PROVISIONING,
		}, nil
	}

	switch resp.Status {
	case sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR:
		status = instancesv1.Status_STATUS_ERROR
	case sharedv1.ResponseStatus_RESPONSE_STATUS_UNSPECIFIED:
		status = instancesv1.Status_STATUS_UNKNOWN
	case sharedv1.ResponseStatus_RESPONSE_STATUS_OK:
		prResp := resp.Response.GetInstanceProvision()
		if prResp == nil {
			status = instancesv1.Status_STATUS_UNKNOWN
		} else {
			status = instancesv1.Status_STATUS_ACTIVE
		}
	}

	return &instancesv1.GetStatusResponse{
		Status: status,
	}, nil
}
