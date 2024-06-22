package runner

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	contextv1 "github.com/powertoolsdev/mono/pkg/types/components/context/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
)

const (
	defaultRunnerTagName   string = "waypoint-runner"
	defaultRunnerTagValue  string = "runner-component"
	defaultRunnerIDTagName string = "runner-id"
)

func (w *wkflow) installECSRunner(ctx workflow.Context, req *runnerv1.ProvisionRunnerRequest) error {
	orgServerAddr := client.DefaultOrgServerAddress(w.cfg.OrgServerRootDomain, req.OrgId)

	// to be able to access the runner, we assume the delegation role, to assume the runner role, to assume the
	// runner install role
	auth := &credentials.Config{
		Region: req.Region,
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:                req.EcsClusterInfo.InstallIamRoleArn,
			SessionName:            fmt.Sprintf("%s-runner-install", req.InstallId),
			SessionDurationSeconds: 60 * 60,

			TwoStepConfig: &assumerole.TwoStepConfig{
				IAMRoleARN: req.AwsSettings.AwsRoleArn,

				SrcIAMRoleARN: req.AwsSettings.AwsRoleDelegation.IamRoleArn,

				// NOTE: static creds are only used for gov-cloud installs
				SrcStaticCredentials: assumerole.StaticCredentials{
					AccessKeyID:     req.AwsSettings.AwsRoleDelegation.AccessKeyId,
					SecretAccessKey: req.AwsSettings.AwsRoleDelegation.SecretAccessKey,
				},
			},
		},
	}

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
		InstallID: req.InstallId,
		Region:    req.Region,
		Auth:      auth,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateEFS, efsReq, &efsResp); err != nil {
		return fmt.Errorf("unable to create efs: %w", err)
	}

	// poll efs
	var pollEFSResp PollEFSResponse
	pollEFSReq := PollEFSRequest{
		InstallID: req.InstallId,
		Region:    req.Region,
		Auth:      auth,
	}
	if err := w.execAWSActivity(ctx, w.act.PollEFS, pollEFSReq, &pollEFSResp); err != nil {
		return fmt.Errorf("unable to poll efs: %w", err)
	}

	// create efs mount targets
	var createEFSMountTargetsResp CreateEFSMountTargetsResponse
	createEFSMountTargetsReq := CreateEFSMountTargetsRequest{
		FsID:   pollEFSResp.FsID,
		Region: req.Region,

		VPCID:           req.EcsClusterInfo.VpcId,
		SubnetIDs:       req.EcsClusterInfo.SubnetIds,
		SecurityGroupID: req.EcsClusterInfo.SecurityGroupId,
		Auth:            auth,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateEFSMountTargets, createEFSMountTargetsReq, &createEFSMountTargetsResp); err != nil {
		return fmt.Errorf("unable to create efs mount targets: %w", err)
	}

	// create efs access mounts
	var createEFSAccessPointsResp CreateEFSAccessPointsResponse
	createEFSAccessPointsReq := CreateEFSAccessPointsRequest{
		InstallID: req.InstallId,
		FsID:      pollEFSResp.FsID,
		Region:    req.Region,
		VPCID:     req.EcsClusterInfo.VpcId,
		SubnetIDs: req.EcsClusterInfo.SubnetIds,
		Auth:      auth,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateEFSAccessPoints, createEFSAccessPointsReq, &createEFSAccessPointsResp); err != nil {
		return fmt.Errorf("unable to create efs access points: %w", err)
	}

	// poll mount targets to be ready
	var pollMountTargetsResp PollEFSMountTargetsResponse
	pollMountTargetsReq := PollEFSMountTargetsRequest{
		FsID:   pollEFSResp.FsID,
		Region: req.Region,
		Auth:   auth,
	}
	if err := w.execAWSActivity(ctx, w.act.PollEFSMountTargets, pollMountTargetsReq, &pollMountTargetsResp); err != nil {
		return fmt.Errorf("unable to create efs access points: %w", err)
	}

	// create log group
	var createLogGroupResp CreateCloudwatchLogGroupResponse
	createLogGroupReq := CreateCloudwatchLogGroupRequest{
		LogGroupName: fmt.Sprintf("waypoint-runner-%s", req.InstallId),
		Region:       req.Region,
		Auth:         auth,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateCloudwatchLogGroup, createLogGroupReq, &createLogGroupResp); err != nil {
		return fmt.Errorf("unable to create cloud watch log group: %w", err)
	}

	// create task definition
	var createTaskDefResp CreateECSTaskDefinitionResponse
	createTaskDefReq := CreateECSTaskDefinitionRequest{
		InstallID: req.InstallId,

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
		Auth:          auth,
	}
	if err := w.execAWSActivity(ctx, w.act.CreateECSTaskDefinition, createTaskDefReq, &createTaskDefResp); err != nil {
		return fmt.Errorf("unable to create task definition: %w", err)
	}

	// create ecs service
	var createServiceResp CreateECSServiceResponse
	createServiceReq := CreateECSServiceRequest{
		ClusterARN: req.EcsClusterInfo.ClusterArn,
		InstallID:  req.InstallId,
		Region:     req.Region,

		SecurityGroupID:   req.EcsClusterInfo.SecurityGroupId,
		SubnetIDs:         req.EcsClusterInfo.SubnetIds,
		TaskDefinitionARN: createTaskDefResp.TaskDefinitionARN,

		Auth: auth,
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

	cwrpReq := CreateWaypointRunnerProfileRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		ClusterInfo:          w.clusterInfo,
		OrgServerAddr:        orgServerAddr,
		InstallID:            req.InstallId,
		OrgID:                req.OrgId,
		AwsRegion:            req.Region,
		RunnerType:           contextv1.RunnerType_RUNNER_TYPE_AWS_ECS,
		// specific fields for ECS runners
		LogGroupName:   fmt.Sprintf("waypoint-runner-%s", req.InstallId),
		EcsClusterInfo: req.EcsClusterInfo,
	}
	_, err = w.createWaypointRunnerProfile(ctx, cwrpReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint runner profile: %w", err)
		return err
	}

	return nil
}

func (w *wkflow) uninstallECSRunner(ctx workflow.Context, req *runnerv1.DeprovisionRunnerRequest) error {
	// to be able to access the runner, we assume the delegation role, to assume the runner role, to assume the
	// runner install role
	auth := &credentials.Config{
		Region: req.Region,
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:                req.EcsClusterInfo.InstallIamRoleArn,
			SessionName:            fmt.Sprintf("%s-runner-install", req.InstallId),
			SessionDurationSeconds: 60 * 60,

			TwoStepConfig: &assumerole.TwoStepConfig{
				IAMRoleARN: req.AwsSettings.AwsRoleArn,

				SrcIAMRoleARN: req.AwsSettings.AwsRoleDelegation.IamRoleArn,
				// NOTE: static creds are only used for gov-cloud installs
				SrcStaticCredentials: assumerole.StaticCredentials{
					AccessKeyID:     req.AwsSettings.AwsRoleDelegation.AccessKeyId,
					SecretAccessKey: req.AwsSettings.AwsRoleDelegation.SecretAccessKey,
				},
			},
		},
	}

	// poll mount targets to be ready
	var deleteServiceResp DeleteServiceResponse
	deleteServiceReq := DeleteServiceRequest{
		InstallID:  req.InstallId,
		ClusterARN: req.EcsClusterInfo.ClusterArn,
		Region:     req.Region,
		Auth:       auth,
	}
	if err := w.execAWSActivity(ctx, w.act.DeleteECSService, deleteServiceReq, &deleteServiceResp); err != nil {
		return fmt.Errorf("unable to delete service: %w", err)
	}

	// poll that service was deleted
	var pollDeleteServiceResp PollDeleteECSServiceResponse
	pollDeleteServiceReq := PollDeleteECSServiceRequest{
		InstallID:  req.InstallId,
		ClusterARN: req.EcsClusterInfo.ClusterArn,
		Region:     req.Region,
		Auth:       auth,
	}
	if err := w.execAWSActivity(ctx, w.act.PollDeleteService, pollDeleteServiceReq, &pollDeleteServiceResp); err != nil {
		return fmt.Errorf("unable to poll deleting ecs service: %w", err)
	}

	var deleteCloudwatchLogGroupResp DeleteCloudwatchLogGroupResponse
	deleteCloudwatchLogGroupReq := DeleteCloudwatchLogGroupRequest{
		LogGroupName: fmt.Sprintf("waypoint-runner-%s", req.InstallId),
		Region:       req.Region,
		Auth:         auth,
	}
	if err := w.execAWSActivity(ctx, w.act.DeleteCloudwatchLogGroup, deleteCloudwatchLogGroupReq, &deleteCloudwatchLogGroupResp); err != nil {
		return fmt.Errorf("unable to poll deleting ecs service: %w", err)
	}

	var deleteEFSAccessPointsResp DeleteEFSAccessPointsResponse
	deleteEFSAccessPointsReq := DeleteEFSAccessPointsRequest{
		InstallID: req.InstallId,
		Region:    req.Region,
		Auth:      auth,
	}
	if err := w.execAWSActivity(ctx, w.act.DeleteEFSAccessPoints, deleteEFSAccessPointsReq, &deleteEFSAccessPointsResp); err != nil {
		return fmt.Errorf("unable to poll deleting ecs service: %w", err)
	}

	var deleteEFSMountTargetsResp DeleteEFSMountTargetsResponse
	deleteEFSMountTargetsReq := DeleteEFSMountTargetsRequest{
		InstallID: req.InstallId,
		Region:    req.Region,
		Auth:      auth,
	}
	if err := w.execAWSActivity(ctx, w.act.DeleteEFSMountTargets, deleteEFSMountTargetsReq, &deleteEFSMountTargetsResp); err != nil {
		return fmt.Errorf("unable to poll deleting ecs service: %w", err)
	}

	var deleteEFSResp DeleteEFSResponse
	deleteEFSReq := DeleteEFSRequest{
		InstallID: req.InstallId,
		Region:    req.Region,
		Auth:      auth,
	}
	if err := w.execAWSActivity(ctx, w.act.DeleteEFS, deleteEFSReq, &deleteEFSResp); err != nil {
		return fmt.Errorf("unable to delete efs: %w", err)
	}

	return nil
}
