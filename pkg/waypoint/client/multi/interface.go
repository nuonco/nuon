package multi

import (
	"context"

	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Client interface {
	// GetVersionInfo returns information about the server. This RPC call does
	// NOT require authentication. It can be used by clients to determine if they
	// are capable of talking to this server.
	GetVersionInfo(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.GetVersionInfoResponse, error)
	// List the available OIDC providers for authentication. The "name" of the
	// OIDC provider can be used with GetOIDCAuthURL and CompleteOIDCAuth to
	// perform OIDC-based authentication.
	ListOIDCAuthMethods(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListOIDCAuthMethodsResponse, error)
	// Get the URL to visit to start authentication with OIDC.
	GetOIDCAuthURL(ctx context.Context, id string, in *pb.GetOIDCAuthURLRequest, opts ...grpc.CallOption) (*pb.GetOIDCAuthURLResponse, error)
	// Complete the OIDC auth cycle after receiving the callback from the
	// OIDC provider.
	CompleteOIDCAuth(ctx context.Context, id string, in *pb.CompleteOIDCAuthRequest, opts ...grpc.CallOption) (*pb.CompleteOIDCAuthResponse, error)
	// Attempts to run a trigger given a trigger ID reference. If the trigger does
	// not exist, we return not found. If the trigger exists but requires authentication
	// we return an error.
	NoAuthRunTrigger(ctx context.Context, id string, in *pb.RunTriggerRequest, opts ...grpc.CallOption) (*pb.RunTriggerResponse, error)
	// GetUser returns the current logged in user or some other user.
	GetUser(ctx context.Context, id string, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.GetUserResponse, error)
	// List all users in the system.
	ListUsers(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListUsersResponse, error)
	// Update the details about an existing user.
	UpdateUser(ctx context.Context, id string, in *pb.UpdateUserRequest, opts ...grpc.CallOption) (*pb.UpdateUserResponse, error)
	// Delete a user. This will invalidate all authentication for this user
	// as well since they no longer exist.
	DeleteUser(ctx context.Context, id string, in *pb.DeleteUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// UpsertAuthMethod upserts the auth method. All users logged in with
	// this auth method will remain logged in even if settings change.
	UpsertAuthMethod(ctx context.Context, id string, in *pb.UpsertAuthMethodRequest, opts ...grpc.CallOption) (*pb.UpsertAuthMethodResponse, error)
	// GetAuthMethod returns the auth method.
	GetAuthMethod(ctx context.Context, id string, in *pb.GetAuthMethodRequest, opts ...grpc.CallOption) (*pb.GetAuthMethodResponse, error)
	// ListAuthMethods returns a list of all the auth methods.
	ListAuthMethods(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListAuthMethodsResponse, error)
	// Delete an auth method. This will invalidate all users authenticated
	// using this auth method and they will have to reauthenticate some other
	// way.
	DeleteAuthMethod(ctx context.Context, id string, in *pb.DeleteAuthMethodRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// ListWorkspaces returns a list of all workspaces.
	//
	// Note that currently this list is never pruned, even if a workspace is
	// no longer in use. We plan to prune this in a future improvement.
	ListWorkspaces(ctx context.Context, id string, in *pb.ListWorkspacesRequest, opts ...grpc.CallOption) (*pb.ListWorkspacesResponse, error)
	// GetWorkspace returns the workspace.
	GetWorkspace(ctx context.Context, id string, in *pb.GetWorkspaceRequest, opts ...grpc.CallOption) (*pb.GetWorkspaceResponse, error)
	// UpsertWorkspace upserts the workspace. Changes to a Workspace's Projects
	// are ignored at this time.
	UpsertWorkspace(ctx context.Context, id string, in *pb.UpsertWorkspaceRequest, opts ...grpc.CallOption) (*pb.UpsertWorkspaceResponse, error)
	// UpsertProject upserts the project.
	UpsertProject(ctx context.Context, id string, in *pb.UpsertProjectRequest, opts ...grpc.CallOption) (*pb.UpsertProjectResponse, error)
	// GetProject returns the project.
	GetProject(ctx context.Context, id string, in *pb.GetProjectRequest, opts ...grpc.CallOption) (*pb.GetProjectResponse, error)
	// ListProjects returns a list of all the projects. There is no equivalent
	// ListApplications because applications are a part of projects and you
	// can use GetProject to get more information about the project.
	ListProjects(ctx context.Context, id string, in *pb.ListProjectsRequest, opts ...grpc.CallOption) (*pb.ListProjectsResponse, error)
	// DestroyProject deletes a project from the database as well as (optionally)
	// destroys all resources created within a project
	DestroyProject(ctx context.Context, id string, in *pb.DestroyProjectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// GetApplication returns one application on the project.
	GetApplication(ctx context.Context, id string, in *pb.GetApplicationRequest, opts ...grpc.CallOption) (*pb.GetApplicationResponse, error)
	// UpsertApplication upserts an application with a project.
	UpsertApplication(ctx context.Context, id string, in *pb.UpsertApplicationRequest, opts ...grpc.CallOption) (*pb.UpsertApplicationResponse, error)
	// ListBuilds returns the builds.
	ListBuilds(ctx context.Context, id string, in *pb.ListBuildsRequest, opts ...grpc.CallOption) (*pb.ListBuildsResponse, error)
	// GetBuild returns a build
	GetBuild(ctx context.Context, id string, in *pb.GetBuildRequest, opts ...grpc.CallOption) (*pb.Build, error)
	// GetLatestBuild returns the most recent successfully completed build
	// for an app.
	GetLatestBuild(ctx context.Context, id string, in *pb.GetLatestBuildRequest, opts ...grpc.CallOption) (*pb.Build, error)
	// ListPushedArtifacts returns the builds.
	ListPushedArtifacts(ctx context.Context, id string, in *pb.ListPushedArtifactsRequest, opts ...grpc.CallOption) (*pb.ListPushedArtifactsResponse, error)
	// GetPushedArtifact returns a deployment
	GetPushedArtifact(ctx context.Context, id string, in *pb.GetPushedArtifactRequest, opts ...grpc.CallOption) (*pb.PushedArtifact, error)
	// GetLatestPushedArtifact returns the most recent successfully completed
	// artifact push for an app.
	GetLatestPushedArtifact(ctx context.Context, id string, in *pb.GetLatestPushedArtifactRequest, opts ...grpc.CallOption) (*pb.PushedArtifact, error)
	// ListDeployments returns the deployments.
	ListDeployments(ctx context.Context, id string, in *pb.ListDeploymentsRequest, opts ...grpc.CallOption) (*pb.ListDeploymentsResponse, error)
	// GetDeployment returns a deployment
	GetDeployment(ctx context.Context, id string, in *pb.GetDeploymentRequest, opts ...grpc.CallOption) (*pb.Deployment, error)
	// ListInstances returns the running instances of deployments.
	ListInstances(ctx context.Context, id string, in *pb.ListInstancesRequest, opts ...grpc.CallOption) (*pb.ListInstancesResponse, error)
	// ListReleases returns the releases.
	ListReleases(ctx context.Context, id string, in *pb.ListReleasesRequest, opts ...grpc.CallOption) (*pb.ListReleasesResponse, error)
	// GetRelease returns a release
	GetRelease(ctx context.Context, id string, in *pb.GetReleaseRequest, opts ...grpc.CallOption) (*pb.Release, error)
	// GetLatestRelease returns the most recent successfully completed
	// release for an app.
	GetLatestRelease(ctx context.Context, id string, in *pb.GetLatestReleaseRequest, opts ...grpc.CallOption) (*pb.Release, error)
	// GetStatusReport returns a StatusReport
	GetStatusReport(ctx context.Context, id string, in *pb.GetStatusReportRequest, opts ...grpc.CallOption) (*pb.StatusReport, error)
	// GetLatestStatusReport returns the most recent successfully completed
	// health report for an app
	GetLatestStatusReport(ctx context.Context, id string, in *pb.GetLatestStatusReportRequest, opts ...grpc.CallOption) (*pb.StatusReport, error)
	// ListStatusReports returns the deployments.
	ListStatusReports(ctx context.Context, id string, in *pb.ListStatusReportsRequest, opts ...grpc.CallOption) (*pb.ListStatusReportsResponse, error)
	// ExpediteStatusReport returns the queued status report job id
	ExpediteStatusReport(ctx context.Context, id string, in *pb.ExpediteStatusReportRequest, opts ...grpc.CallOption) (*pb.ExpediteStatusReportResponse, error)
	// GetLogStream reads the log stream for a deployment. This will immediately
	// send a single LogEntry with the lines we have so far. If there are no
	// available lines this will NOT block and instead will return an error.
	// The client can choose to retry or not.
	GetLogStream(ctx context.Context, id string, in *pb.GetLogStreamRequest, opts ...grpc.CallOption) (pb.Waypoint_GetLogStreamClient, error)
	// StartExecStream starts an exec session.
	StartExecStream(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_StartExecStreamClient, error)
	// Set one or more configuration variables for applications or runners.
	SetConfig(ctx context.Context, id string, in *pb.ConfigSetRequest, opts ...grpc.CallOption) (*pb.ConfigSetResponse, error)
	// Retrieve merged configuration values for a specific scope. You can determine
	// where a configuration variable was set by looking at the scope field on
	// each variable.
	GetConfig(ctx context.Context, id string, in *pb.ConfigGetRequest, opts ...grpc.CallOption) (*pb.ConfigGetResponse, error)
	// Set the configuration for a dynamic configuration source. If you're looking
	// to set application configuration, you probably want SetConfig instead.
	SetConfigSource(ctx context.Context, id string, in *pb.SetConfigSourceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Get the matching configuration source for the request. This will return
	// the most specific matching config source given the scope in the request.
	// For example, if you search for an app-specific config source and only
	// a global config exists, the global config will be returned.
	GetConfigSource(ctx context.Context, id string, in *pb.GetConfigSourceRequest, opts ...grpc.CallOption) (*pb.GetConfigSourceResponse, error)
	// Create a hostname with the URL service.
	CreateHostname(ctx context.Context, id string, in *pb.CreateHostnameRequest, opts ...grpc.CallOption) (*pb.CreateHostnameResponse, error)
	// Delete a hostname with the URL service.
	DeleteHostname(ctx context.Context, id string, in *pb.DeleteHostnameRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// List all our registered hostnames.
	ListHostnames(ctx context.Context, id string, in *pb.ListHostnamesRequest, opts ...grpc.CallOption) (*pb.ListHostnamesResponse, error)
	// QueueJob queues a job for execution by a runner. This will return as
	// soon as the job is queued, it will not wait for execution.
	QueueJob(ctx context.Context, id string, in *pb.QueueJobRequest, opts ...grpc.CallOption) (*pb.QueueJobResponse, error)
	// CancelJob cancels a job. If the job is still queued this is a quick
	// and easy operation. If the job is already completed, then this does
	// nothing. If the job is assigned or running, then this will signal
	// the runner about the cancellation but it may take time.
	//
	// This RPC always returns immediately. You must use GetJob or GetJobStream
	// to wait on the status of the cancellation.
	CancelJob(ctx context.Context, id string, in *pb.CancelJobRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// GetJob queries a job by ID.
	GetJob(ctx context.Context, id string, in *pb.GetJobRequest, opts ...grpc.CallOption) (*pb.Job, error)
	// ListJobs will return a list of jobs known to Waypoint server. Can be filtered
	// by request on values like workspace, project, application, job state, etc.
	ListJobs(ctx context.Context, id string, in *pb.ListJobsRequest, opts ...grpc.CallOption) (*pb.ListJobsResponse, error)
	// ValidateJob checks if a job appears valid. This will check the job
	// structure itself (i.e. missing fields) and can also check to ensure
	// the job is assignable to a runner.
	ValidateJob(ctx context.Context, id string, in *pb.ValidateJobRequest, opts ...grpc.CallOption) (*pb.ValidateJobResponse, error)
	// GetJobStream opens a job event stream for a running job. This can be
	// used to listen for terminal output and other events of a running job.
	// Multiple listeners can open a job stream.
	GetJobStream(ctx context.Context, id string, in *pb.GetJobStreamRequest, opts ...grpc.CallOption) (pb.Waypoint_GetJobStreamClient, error)
	// GetRunner gets information about a single runner.
	GetRunner(ctx context.Context, id string, in *pb.GetRunnerRequest, opts ...grpc.CallOption) (*pb.Runner, error)
	// ListRunners lists runners that are currently registered with the waypoint server.
	// This list does not include previous on-demand runners that have exited.
	ListRunners(ctx context.Context, id string, in *pb.ListRunnersRequest, opts ...grpc.CallOption) (*pb.ListRunnersResponse, error)
	// AdoptRunners allows marking a runner as adopted or rejected.
	AdoptRunner(ctx context.Context, id string, in *pb.AdoptRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// ForgetRunner deletes an existing runner entry and makes the server
	// behave as if the runner no longer exists. If the runner is currently
	// running, it will receive errors on subsequent jobs, and will have to
	// re-register. A forgotten runner will not be assigned new jobs until
	// re-registered.
	ForgetRunner(ctx context.Context, id string, in *pb.ForgetRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// GetServerConfig sets configuration for the Waypoint server.
	GetServerConfig(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.GetServerConfigResponse, error)
	// SetServerConfig sets configuration for the Waypoint server.
	SetServerConfig(ctx context.Context, id string, in *pb.SetServerConfigRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// CreateSnapshot creates a new database snapshot.
	CreateSnapshot(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (pb.Waypoint_CreateSnapshotClient, error)
	// RestoreSnapshot performs a database restore with the given snapshot.
	// This API doesn't do a full online restore, it only stages the restore
	// for the next server start to finalize the restore. See the arguments for
	// more information.
	RestoreSnapshot(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_RestoreSnapshotClient, error)
	// BootstrapToken returns the initial token for the server. This can only
	// be requested once on first startup. After initial request this will
	// always return a PermissionDenied error.
	BootstrapToken(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.NewTokenResponse, error)
	// DecodeToken takes a token string and returns the structured information
	// about the given token. This is useful for frontends (CLI, UI, etc.) to
	// learn more about a token before using it. For example, if a UI wants to
	// create a signup flow around signup tokens, they can validate the token
	// ahead of time.
	//
	// This endpoint does NOT require authentication.
	DecodeToken(ctx context.Context, id string, in *pb.DecodeTokenRequest, opts ...grpc.CallOption) (*pb.DecodeTokenResponse, error)
	// Generate a new invite token that users can exchange for a login token.
	// This can be used to also invite new users to the Waypoint server.
	GenerateInviteToken(ctx context.Context, id string, in *pb.InviteTokenRequest, opts ...grpc.CallOption) (*pb.NewTokenResponse, error)
	// Generate a new login token that users can use to login directly.
	// This can only be called for existing users.
	GenerateLoginToken(ctx context.Context, id string, in *pb.LoginTokenRequest, opts ...grpc.CallOption) (*pb.NewTokenResponse, error)
	// Generate a new runner token that can be used with runners so they
	// immediately begin work. The recommended appraoch is to instead use
	// the adoption flow but this also works.
	GenerateRunnerToken(ctx context.Context, id string, in *pb.GenerateRunnerTokenRequest, opts ...grpc.CallOption) (*pb.NewTokenResponse, error)
	// Exchange a invite token for a login token. If the invite token is
	// for a new user, this will create a new user account with the provided
	// username hint.
	ConvertInviteToken(ctx context.Context, id string, in *pb.ConvertInviteTokenRequest, opts ...grpc.CallOption) (*pb.NewTokenResponse, error)
	// RunnerToken is called to register a runner and request a token for
	// remaining runner API calls. This kicks off the "adoption" process
	// (if necessary).
	//
	// This is unauthenticated (but requires a cookie in the metadata).
	RunnerToken(ctx context.Context, id string, in *pb.RunnerTokenRequest, opts ...grpc.CallOption) (*pb.RunnerTokenResponse, error)
	// RunnerConfig is called to receive the configuration for the runner.
	// The response is a stream so that the configuration can be updated later.
	RunnerConfig(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_RunnerConfigClient, error)
	// RunnerJobStream is called by a runner to request a single job for
	// execution and update the status of that job.
	RunnerJobStream(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_RunnerJobStreamClient, error)
	// RunnerGetDeploymentConfig is called by a runner for a deployment operation
	// to determine the settings to use for a deployment.
	RunnerGetDeploymentConfig(ctx context.Context, id string, in *pb.RunnerGetDeploymentConfigRequest, opts ...grpc.CallOption) (*pb.RunnerGetDeploymentConfigResponse, error)
	// EntrypointConfig is called to get the configuration for the entrypoint
	// and also to get any potential updates.
	//
	// This endpoint also registers the instance with the server. This MUST be
	// called first otherwise other RPCs related to the entrypoint may fail
	// with FailedPrecondition.
	EntrypointConfig(ctx context.Context, id string, in *pb.EntrypointConfigRequest, opts ...grpc.CallOption) (pb.Waypoint_EntrypointConfigClient, error)
	// EntrypointLogStream is called to open the stream that logs are sent to.
	EntrypointLogStream(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_EntrypointLogStreamClient, error)
	// EntrypointExecStream is called to open the data stream for the exec session.
	EntrypointExecStream(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_EntrypointExecStreamClient, error)
	// WaypointHclFmt formats a waypoint.hcl file. This must be in HCL format.
	// JSON formatting is not supported.
	WaypointHclFmt(ctx context.Context, id string, in *pb.WaypointHclFmtRequest, opts ...grpc.CallOption) (*pb.WaypointHclFmtResponse, error)
	// UpsertOnDemandRunnerConfig updates or inserts a on-demand runner
	// configuration. This configuration can be used by projects for running
	// operations on just-in-time launched runners.
	UpsertOnDemandRunnerConfig(ctx context.Context, id string, in *pb.UpsertOnDemandRunnerConfigRequest, opts ...grpc.CallOption) (*pb.UpsertOnDemandRunnerConfigResponse, error)
	// GetOnDemandRunnerConfig returns the on-demand runner configuration.
	GetOnDemandRunnerConfig(ctx context.Context, id string, in *pb.GetOnDemandRunnerConfigRequest, opts ...grpc.CallOption) (*pb.GetOnDemandRunnerConfigResponse, error)
	// GetOnDemandRunnerConfig returns the on-demand runner configuration.
	GetDefaultOnDemandRunnerConfig(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.GetOnDemandRunnerConfigResponse, error)
	// GetOnDemandRunnerConfig returns the on-demand runner configuration.
	DeleteOnDemandRunnerConfig(ctx context.Context, id string, in *pb.DeleteOnDemandRunnerConfigRequest, opts ...grpc.CallOption) (*pb.DeleteOnDemandRunnerConfigResponse, error)
	// ListOnDemandRunnerConfigs returns a list of all the on-demand runners configs.
	ListOnDemandRunnerConfigs(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListOnDemandRunnerConfigsResponse, error)
	// UpsertBuild updates or inserts a build. A build is responsible for
	// taking some set of source information and turning it into an initial
	// artifact. This artifact is considered "local" until it is pushed.
	UpsertBuild(ctx context.Context, id string, in *pb.UpsertBuildRequest, opts ...grpc.CallOption) (*pb.UpsertBuildResponse, error)
	// UpsertPushedArtifact updates or inserts a pushed artifact. This is
	// useful for local operations to work on a pushed artifact.
	UpsertPushedArtifact(ctx context.Context, id string, in *pb.UpsertPushedArtifactRequest, opts ...grpc.CallOption) (*pb.UpsertPushedArtifactResponse, error)
	// UpsertDeployment updates or inserts a deployment.
	UpsertDeployment(ctx context.Context, id string, in *pb.UpsertDeploymentRequest, opts ...grpc.CallOption) (*pb.UpsertDeploymentResponse, error)
	// UpsertRelease updates or inserts a release.
	UpsertRelease(ctx context.Context, id string, in *pb.UpsertReleaseRequest, opts ...grpc.CallOption) (*pb.UpsertReleaseResponse, error)
	// UpsertStatusReport updates or inserts a statusreport.
	UpsertStatusReport(ctx context.Context, id string, in *pb.UpsertStatusReportRequest, opts ...grpc.CallOption) (*pb.UpsertStatusReportResponse, error)
	// GetTask returns a requested Task message. Or an error if it does not exist.
	GetTask(ctx context.Context, id string, in *pb.GetTaskRequest, opts ...grpc.CallOption) (*pb.GetTaskResponse, error)
	// ListTask will return a list of all existing Tasks
	ListTask(ctx context.Context, id string, in *pb.ListTaskRequest, opts ...grpc.CallOption) (*pb.ListTaskResponse, error)
	// CancelTask will attempt to gracefully cancel each job in the task job triple
	CancelTask(ctx context.Context, id string, in *pb.CancelTaskRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// UpsertTrigger updates or inserts a trigger URL configuration.
	UpsertTrigger(ctx context.Context, id string, in *pb.UpsertTriggerRequest, opts ...grpc.CallOption) (*pb.UpsertTriggerResponse, error)
	// GetTrigger returns a requested trigger message. Or an error if it does not exist.
	GetTrigger(ctx context.Context, id string, in *pb.GetTriggerRequest, opts ...grpc.CallOption) (*pb.GetTriggerResponse, error)
	// DeleteTrigger takes a trigger id and deletes it, if it exists.
	DeleteTrigger(ctx context.Context, id string, in *pb.DeleteTriggerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// ListTriggers takes a request filter, and returns any matching existing triggers
	ListTriggers(ctx context.Context, id string, in *pb.ListTriggerRequest, opts ...grpc.CallOption) (*pb.ListTriggerResponse, error)
	// RunTrigger will look up the referenced trigger and attempt to queue a job
	// based on the trigger configuration.
	RunTrigger(ctx context.Context, id string, in *pb.RunTriggerRequest, opts ...grpc.CallOption) (*pb.RunTriggerResponse, error)
	// UpsertPipeline updates or inserts a pipeline. This is an INTERNAL ONLY
	// endpoint that is meant to only be called by runners. Calling this manually
	// can risk the internal state for pipelines. In the future, we'll restrict
	// access to this via ACLs.
	UpsertPipeline(ctx context.Context, id string, in *pb.UpsertPipelineRequest, opts ...grpc.CallOption) (*pb.UpsertPipelineResponse, error)
	// RunPipeline queues a pipeline execution.
	RunPipeline(ctx context.Context, id string, in *pb.RunPipelineRequest, opts ...grpc.CallOption) (*pb.RunPipelineResponse, error)
	// GetPipeline returns a pipeline proto by pipeline ref id
	GetPipeline(ctx context.Context, id string, in *pb.GetPipelineRequest, opts ...grpc.CallOption) (*pb.GetPipelineResponse, error)
	// GetPipelineRun returns a pipeline run proto by pipeline ref id and sequence
	GetPipelineRun(ctx context.Context, id string, in *pb.GetPipelineRunRequest, opts ...grpc.CallOption) (*pb.GetPipelineRunResponse, error)
	// GetLatestPipelineRun returns a pipeline run proto by pipeline ref id and sequence
	GetLatestPipelineRun(ctx context.Context, id string, in *pb.GetPipelineRequest, opts ...grpc.CallOption) (*pb.GetPipelineRunResponse, error)
	// ListPipelines takes a project and evaluates the projects config to get
	// a list of Pipeline protos to return in the response. These pipelines
	// are scoped to a single project from the request. It will return an
	// error if the requested project does not exist, or an empty response
	// if no pipelines are defined for the project.
	ListPipelines(ctx context.Context, id string, in *pb.ListPipelinesRequest, opts ...grpc.CallOption) (*pb.ListPipelinesResponse, error)
	// ListPipelineRuns takes a pipeline ref and returns a list of runs of that pipeline.
	// It will return an error if the requested pipeline does not exist, or an empty response
	// if there are no runs for the pipeline.
	ListPipelineRuns(ctx context.Context, id string, in *pb.ListPipelineRunsRequest, opts ...grpc.CallOption) (*pb.ListPipelineRunsResponse, error)
	// ConfigSyncPipeline takes a request for a given project and syncs the current
	// project config to the Waypoint database.
	ConfigSyncPipeline(ctx context.Context, id string, in *pb.ConfigSyncPipelineRequest, opts ...grpc.CallOption) (*pb.ConfigSyncPipelineResponse, error)
	// List full projects (not just refs)
	UI_ListProjects(ctx context.Context, id string, in *pb.UI_ListProjectsRequest, opts ...grpc.CallOption) (*pb.UI_ListProjectsResponse, error)
	// Get a given project with useful related records.
	UI_GetProject(ctx context.Context, id string, in *pb.UI_GetProjectRequest, opts ...grpc.CallOption) (*pb.UI_GetProjectResponse, error)
	// List deployments for a given application.
	UI_ListDeployments(ctx context.Context, id string, in *pb.UI_ListDeploymentsRequest, opts ...grpc.CallOption) (*pb.UI_ListDeploymentsResponse, error)
	// GetDeployment returns a deployment
	UI_GetDeployment(ctx context.Context, id string, in *pb.UI_GetDeploymentRequest, opts ...grpc.CallOption) (*pb.UI_GetDeploymentResponse, error)
	// List releases for a given application.
	UI_ListReleases(ctx context.Context, id string, in *pb.UI_ListReleasesRequest, opts ...grpc.CallOption) (*pb.UI_ListReleasesResponse, error)
}
