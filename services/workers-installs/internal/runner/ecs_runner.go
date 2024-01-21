package runner

import (
	"fmt"

	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultRunnerTagName   string = "waypoint-runner"
	defaultRunnerTagValue  string = "runner-component"
	defaultRunnerIDTagName string = "runner-id"
)

func (w *wkflow) installECSRunner(ctx workflow.Context, req *runnerv1.ProvisionRunnerRequest) error {
	orgServerAddr := client.DefaultOrgServerAddress(w.cfg.OrgServerRootDomain, req.OrgId)

	// get waypoint server cookie
	gwscReq := GetWaypointServerCookieRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		ClusterInfo:          w.clusterInfo,
	}
	gwscResp, err := w.getWaypointServerCookie(ctx, gwscReq)
	if err != nil {
		err = fmt.Errorf("failed to get waypoint server cookie: %w", err)
		return err
	}

	// create efs
	var efsResp CreateEFSResponse
	efsReq := CreateEFSRequest{
		IAMRoleARN: req.EcsClusterInfo.InstallIamRoleArn,
		InstallID:  req.InstallId,
		Region:     req.Region,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateEFS, efsReq, &efsResp); err != nil {
		return fmt.Errorf("unable to create efs: %w", err)
	}

	// poll efs
	var pollEFSResp PollEFSResponse
	pollEFSReq := PollEFSRequest{
		IAMRoleARN: req.EcsClusterInfo.InstallIamRoleArn,
		InstallID:  req.InstallId,
		Region:     req.Region,
	}
	if err := w.execAWSActivity(ctx, w.act.PollEFS, pollEFSReq, &pollEFSResp); err != nil {
		return fmt.Errorf("unable to poll efs: %w", err)
	}

	// create efs mount targets
	var createEFSMountTargetsResp CreateEFSMountTargetsResponse
	createEFSMountTargetsReq := CreateEFSMountTargetsRequest{
		IAMRoleARN: req.EcsClusterInfo.InstallIamRoleArn,
		FsID:       pollEFSResp.FsID,
		Region:     req.Region,

		VPCID:           req.EcsClusterInfo.VpcId,
		SubnetIDs:       req.EcsClusterInfo.SubnetIds,
		SecurityGroupID: req.EcsClusterInfo.SecurityGroupId,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateEFSMountTargets, createEFSMountTargetsReq, &createEFSMountTargetsResp); err != nil {
		return fmt.Errorf("unable to create efs mount targets: %w", err)
	}

	// create efs access mounts
	var createEFSAccessPointsResp CreateEFSAccessPointsResponse
	createEFSAccessPointsReq := CreateEFSAccessPointsRequest{
		IAMRoleARN: req.EcsClusterInfo.InstallIamRoleArn,
		FsID:       pollEFSResp.FsID,
		Region:     req.Region,
		VPCID:      req.EcsClusterInfo.VpcId,
		SubnetIDs:  req.EcsClusterInfo.SubnetIds,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateEFSAccessPoints, createEFSAccessPointsReq, &createEFSAccessPointsResp); err != nil {
		return fmt.Errorf("unable to create efs access points: %w", err)
	}

	// poll mount targets to be ready
	var pollMountTargetsResp PollEFSMountTargetsResponse
	pollMountTargetsReq := PollEFSMountTargetsRequest{
		IAMRoleARN: req.EcsClusterInfo.InstallIamRoleArn,
		FsID:       pollEFSResp.FsID,
		Region:     req.Region,
	}
	if err := w.execAWSActivity(ctx, w.act.PollEFSMountTargets, pollMountTargetsReq, &pollMountTargetsResp); err != nil {
		return fmt.Errorf("unable to create efs access points: %w", err)
	}

	// create log group
	var createLogGroupResp CreateCloudwatchLogGroupResponse
	createLogGroupReq := CreateCloudwatchLogGroupRequest{
		IAMRoleARN:   req.EcsClusterInfo.InstallIamRoleArn,
		LogGroupName: fmt.Sprintf("waypoint-runner-%s", req.InstallId),
		Region:       req.Region,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateCloudwatchLogGroup, createLogGroupReq, &createLogGroupResp); err != nil {
		return fmt.Errorf("unable to create cloud watch log group: %w", err)
	}

	// create task definition
	var createTaskDefResp CreateECSTaskDefinitionResponse
	createTaskDefReq := CreateECSTaskDefinitionRequest{
		IAMRoleARN: req.EcsClusterInfo.InstallIamRoleArn,
		InstallID:  req.InstallId,

		RunnerRoleARN: req.EcsClusterInfo.RunnerIamRoleArn,
		EnvVars: map[string]string{
			"WAYPOINT_SERVER_ADDR":            client.DefaultOrgServerAddress(w.cfg.OrgServerRootDomain, req.OrgId),
			"WAYPOINT_SERVER_TLS":             "true",
			"WAYPOINT_SERVER_TLS_SKIP_VERIFY": "true",
		},
		LogGroupName:  createLogGroupResp.LogGroupName,
		Region:        req.Region,
		ServerCookie:  gwscResp.Cookie,
		AccessPointID: createEFSAccessPointsResp.AccessPointIDs[0],
		FileSystemID:  pollEFSResp.FsID,
		Args:          []string{},
	}
	if err := w.execAWSActivity(ctx, w.act.CreateECSTaskDefinition, createTaskDefReq, &createTaskDefResp); err != nil {
		return fmt.Errorf("unable to create task definition: %w", err)
	}

	// create ecs service
	var createServiceResp CreateECSServiceResponse
	createServiceReq := CreateECSServiceRequest{
		IAMRoleARN: req.EcsClusterInfo.InstallIamRoleArn,
		ClusterARN: req.EcsClusterInfo.ClusterArn,
		InstallID:  req.InstallId,
		Region:     req.Region,

		SecurityGroupID:   req.EcsClusterInfo.SecurityGroupId,
		SubnetIDs:         req.EcsClusterInfo.SubnetIds,
		TaskDefinitionARN: createTaskDefResp.TaskDefinitionARN,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateECSService, createServiceReq, &createServiceResp); err != nil {
		return fmt.Errorf("unable to create service: %w", err)
	}

	awrReq := AdoptWaypointRunnerRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          w.clusterInfo,
	}
	_, err = w.adoptWaypointRunner(ctx, awrReq)
	if err != nil {
		err = fmt.Errorf("failed to adopt waypoint runner: %w", err)
		return err
	}

	return nil
}

func (w *wkflow) uninstallECSRunner(ctx workflow.Context, req *runnerv1.DeprovisionRunnerRequest) error {
	// poll mount targets to be ready
	var deleteServiceResp DeleteServiceResponse
	deleteServiceReq := DeleteServiceRequest{
		InstallID:  req.InstallId,
		IAMRoleARN: req.EcsClusterInfo.InstallIamRoleArn,
		ClusterARN: req.EcsClusterInfo.ClusterArn,
		Region:     req.Region,
	}
	if err := w.execAWSActivity(ctx, w.act.DeleteECSService, deleteServiceReq, &deleteServiceResp); err != nil {
		return fmt.Errorf("unable to delete service: %w", err)
	}

	// poll mount targets to be ready
	var pollDeleteServiceResp PollDeleteECSServiceResponse
	pollDeleteServiceReq := PollDeleteECSServiceRequest{
		InstallID:  req.InstallId,
		IAMRoleARN: req.EcsClusterInfo.InstallIamRoleArn,
		ClusterARN: req.EcsClusterInfo.ClusterArn,
		Region:     req.Region,
	}
	if err := w.execAWSActivity(ctx, w.act.PollDeleteService, pollDeleteServiceReq, &pollDeleteServiceResp); err != nil {
		return fmt.Errorf("unable to poll deleting ecs service: %w", err)
	}

	return nil
}
