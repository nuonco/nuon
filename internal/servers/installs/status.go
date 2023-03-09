package installs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	"github.com/powertoolsdev/orgs-api/internal/servers"
	installsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/installs/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

func ensureShortIDsGetStatusRequest(msg *installsv1.GetStatusRequest) error {
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

	installID, err := servers.EnsureShortID(msg.InstallId)
	if err != nil {
		return fmt.Errorf("invalid installID: %w", err)
	}
	msg.InstallId = installID

	return nil
}

func (s *server) GetStatus(
	ctx context.Context,
	req *connect.Request[installsv1.GetStatusRequest],
) (*connect.Response[installsv1.GetStatusResponse], error) {
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
func (s *server) getStatus(ctx context.Context, req *installsv1.GetStatusRequest, wkflows workflows.Repo) (*installsv1.GetStatusResponse, error) {
	var status installsv1.Status

	resp, err := wkflows.GetInstallProvisionResponse(ctx, req.OrgId, req.AppId, req.InstallId)
	if err != nil {
		req, err := wkflows.GetInstallProvisionRequest(ctx, req.OrgId, req.AppId, req.InstallId)
		if err != nil || req.Request.GetInstallProvision() == nil {
			return &installsv1.GetStatusResponse{
				Status: installsv1.Status_STATUS_UNKNOWN,
			}, nil
		}

		return &installsv1.GetStatusResponse{
			Status: installsv1.Status_STATUS_PROVISIONING,
		}, nil
	}

	switch resp.Status {
	case sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR:
		status = installsv1.Status_STATUS_ERROR
	case sharedv1.ResponseStatus_RESPONSE_STATUS_UNSPECIFIED:
		status = installsv1.Status_STATUS_UNKNOWN
	case sharedv1.ResponseStatus_RESPONSE_STATUS_OK:
		prResp := resp.Response.GetInstallProvision()
		if prResp == nil {
			status = installsv1.Status_STATUS_UNKNOWN
		} else {
			status = installsv1.Status_STATUS_ACTIVE
		}
	}

	return &installsv1.GetStatusResponse{
		Status: status,
	}, nil
}
