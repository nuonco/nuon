package services

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func (s *service) GetInfo(ctx context.Context, orgID string) (*orgsv1.GetInfoResponse, error) {
	workflowResp, err := s.WorkflowsRepo.GetOrgProvisionResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow response: %w", err)
	}
	signupWorkflowResp := workflowResp.Response.GetOrgSignup()
	if signupWorkflowResp == nil {
		return nil, fmt.Errorf("invalid workflow response")
	}
	iamRoles := signupWorkflowResp.IamRoles
	wpServer := signupWorkflowResp.Server

	serverInfo, err := s.WaypointRepo.GetVersionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get version info: %w", err)
	}

	runnerInfo, err := s.WaypointRepo.GetRunner(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner: %w", err)
	}

	return &orgsv1.GetInfoResponse{
		Id: orgID,
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
