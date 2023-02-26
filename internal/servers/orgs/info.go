package orgs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func (s *server) GetInfo(
	ctx context.Context,
	req *connect.Request[orgsv1.GetInfoRequest],
) (*connect.Response[orgsv1.GetInfoResponse], error) {
	wkflows, err := s.WorkflowsRepo(ctx, req.Msg.OrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflows repo: %w", err)
	}

	wp, err := s.WaypointRepo(ctx, req.Msg.OrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get waypoint repo: %w", err)
	}

	resp, err := s.getInfo(ctx, req.Msg, wkflows, wp)
	if err != nil {
		return nil, fmt.Errorf("unable to get info: %w", err)
	}
	return connect.NewResponse(resp), nil
}

func (s *server) getInfo(
	ctx context.Context,
	req *orgsv1.GetInfoRequest,
	wkflows workflows.Repo,
	wp waypoint.Repo,
) (*orgsv1.GetInfoResponse, error) {
	workflowResp, err := wkflows.GetOrgProvisionResponse(ctx, req.OrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow response: %w", err)
	}
	signupWorkflowResp := workflowResp.Response.GetOrgSignup()
	if signupWorkflowResp == nil {
		return nil, fmt.Errorf("invalid workflow response")
	}
	iamRoles := signupWorkflowResp.IamRoles
	wpServer := signupWorkflowResp.Server

	serverInfo, err := wp.GetVersionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get version info: %w", err)
	}

	runnerInfo, err := wp.GetRunner(ctx, req.OrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner: %w", err)
	}

	return &orgsv1.GetInfoResponse{
		Id: req.OrgId,
		ServerInfo: &orgsv1.ServerInfo{
			Version:                serverInfo.Info.Version,
			ProtocolVersion:        fmt.Sprintf("%d", serverInfo.Info.Api.Current),
			MinimumProtocolVersion: fmt.Sprintf("%d", serverInfo.Info.Api.Minimum),
			Address:                wpServer.ServerAddress,
			SecretNamespace:        wpServer.SecretName,
		},
		BuildRunnerInfo: &orgsv1.RunnerInfo{
			Id:            runnerInfo.Id,
			Kind:          fmt.Sprintf("%s", runnerInfo.Kind),
			Labels:        runnerInfo.Labels,
			Online:        runnerInfo.Online,
			AdoptionState: runnerInfo.AdoptionState.String(),
			FirstSeen:     waypoint.TimestampToDatetime(runnerInfo.FirstSeen),
			LastSeen:      waypoint.TimestampToDatetime(runnerInfo.LastSeen),
		},
		IamInfo: &orgsv1.IAMInfo{
			Roles: map[string]*orgsv1.IAMRole{
				"deployments": {
					Arn:         iamRoles.DeploymentsRoleArn,
					Description: "IAM role used to manage access to the deployments bucket.",
				},
				"installations": {
					Arn:         iamRoles.InstallationsRoleArn,
					Description: "IAM role used to manage access to the installations bucket.",
				},
				"odr": {
					Arn:         iamRoles.OdrRoleArn,
					Description: "IAM role that the ODR in our account uses to create builds.",
				},
				"instances": {
					Arn:         iamRoles.InstancesRoleArn,
					Description: "IAM role that instances use, to access ECR.",
				},
				"installer": {
					Arn:         iamRoles.InstancesRoleArn,
					Description: "IAM role that is used when doing installations, and should be given access to by end customers.",
				},
				"orgs": {
					Arn:         iamRoles.OrgsRoleArn,
					Description: "IAM role used to managed access to the deployments bucket.",
				},
			},
		},
	}, nil
}
