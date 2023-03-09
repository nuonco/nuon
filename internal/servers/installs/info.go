package installs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	"github.com/powertoolsdev/orgs-api/internal/servers"
	installsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/installs/v1"
)

func ensureShortIDsGetInfoRequest(msg *installsv1.GetInfoRequest) error {
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

func (s *server) GetInfo(
	ctx context.Context,
	req *connect.Request[installsv1.GetInfoRequest],
) (*connect.Response[installsv1.GetInfoResponse], error) {
	if err := ensureShortIDsGetInfoRequest(req.Msg); err != nil {
		return nil, fmt.Errorf("unable to ensure ids: %w", err)
	}

	wkflowsRepo, err := s.WorkflowsRepo(ctx, req.Msg.OrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflows repo: %w", err)
	}

	resp, err := s.getInfo(ctx, req.Msg, wkflowsRepo)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	return connect.NewResponse(resp), nil
}

func (s *server) getInfo(ctx context.Context, req *installsv1.GetInfoRequest, wkflows workflows.Repo) (*installsv1.GetInfoResponse, error) {
	resp, err := wkflows.GetInstallProvisionResponse(ctx, req.OrgId, req.AppId, req.InstallId)
	if err != nil {
		return nil, fmt.Errorf("unable to get install provision response: %w", err)
	}

	prResp := resp.Response.GetInstallProvision()
	if prResp == nil {
		return nil, fmt.Errorf("invalid response object")
	}

	return &installsv1.GetInfoResponse{
		Id:       req.InstallId,
		Response: prResp,
	}, nil
}
