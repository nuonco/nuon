package apps

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	"github.com/powertoolsdev/orgs-api/internal/servers"
	appsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/apps/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

func ensureShortIDsGetStatusRequest(msg *appsv1.GetStatusRequest) error {
	orgID, err := servers.EnsureShortID(msg.OrgId)
	if err != nil {
		return fmt.Errorf("invalid orgID: %w", err)
	}
	msg.OrgId = orgID

	appID, err := servers.EnsureShortID(msg.AppId)
	if err != nil {
		return fmt.Errorf("invalid appID: %w", err)
	}
	msg.AppId = appID

	return nil
}

func (s *server) GetStatus(
	ctx context.Context,
	req *connect.Request[appsv1.GetStatusRequest],
) (*connect.Response[appsv1.GetStatusResponse], error) {
	if err := ensureShortIDsGetStatusRequest(req.Msg); err != nil {
		return nil, fmt.Errorf("unable to ensure ids: %w", err)
	}

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
func (s *server) getStatus(ctx context.Context, req *appsv1.GetStatusRequest, wkflows workflows.Repo) (*appsv1.GetStatusResponse, error) {
	var status appsv1.Status

	resp, err := wkflows.GetAppProvisionResponse(ctx, req.OrgId, req.AppId)
	if err != nil {
		req, err := wkflows.GetAppProvisionRequest(ctx, req.OrgId, req.AppId)
		if err != nil || req.Request.GetAppProvision() == nil {
			return &appsv1.GetStatusResponse{
				Status: appsv1.Status_STATUS_UNKNOWN,
			}, nil
		}

		return &appsv1.GetStatusResponse{
			Status: appsv1.Status_STATUS_PROVISIONING,
		}, nil
	}

	switch resp.Status {
	case sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR:
		status = appsv1.Status_STATUS_ERROR
	case sharedv1.ResponseStatus_RESPONSE_STATUS_UNSPECIFIED:
		status = appsv1.Status_STATUS_UNKNOWN
	case sharedv1.ResponseStatus_RESPONSE_STATUS_OK:
		prResp := resp.Response.GetAppsProvision()
		if prResp == nil {
			status = appsv1.Status_STATUS_UNKNOWN
		} else {
			status = appsv1.Status_STATUS_ACTIVE
		}
	}

	return &appsv1.GetStatusResponse{
		Status: status,
	}, nil
}
