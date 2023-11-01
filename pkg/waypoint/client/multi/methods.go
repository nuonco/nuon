package multi

import (
	"context"
	"fmt"

	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetVersionInfo returns information about the server. This RPC call does
// NOT require authentication. It can be used by clients to determine if they
// are capable of talking to this server.
func (m *multiClient) GetVersionInfo(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.GetVersionInfoResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetVersionInfo(ctx, in, opts...)
}

// List the available OIDC providers for authentication. The "name" of the
// OIDC provider can be used with GetOIDCAuthURL and CompleteOIDCAuth to
// perform OIDC-based authentication.
func (m *multiClient) ListOIDCAuthMethods(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListOIDCAuthMethodsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListOIDCAuthMethods(ctx, in, opts...)
}

// Get the URL to visit to start authentication with OIDC.
func (m *multiClient) GetOIDCAuthURL(ctx context.Context, id string, in *pb.GetOIDCAuthURLRequest, opts ...grpc.CallOption) (*pb.GetOIDCAuthURLResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetOIDCAuthURL(ctx, in, opts...)
}

// Complete the OIDC auth cycle after receiving the callback from the
// OIDC provider.
func (m *multiClient) CompleteOIDCAuth(ctx context.Context, id string, in *pb.CompleteOIDCAuthRequest, opts ...grpc.CallOption) (*pb.CompleteOIDCAuthResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.CompleteOIDCAuth(ctx, in, opts...)
}

// Attempts to run a trigger given a trigger ID reference. If the trigger does
// not exist, we return not found. If the trigger exists but requires authentication
// we return an error.
func (m *multiClient) NoAuthRunTrigger(ctx context.Context, id string, in *pb.RunTriggerRequest, opts ...grpc.CallOption) (*pb.RunTriggerResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.NoAuthRunTrigger(ctx, in, opts...)
}

// GetUser returns the current logged in user or some other user.
func (m *multiClient) GetUser(ctx context.Context, id string, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.GetUserResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetUser(ctx, in, opts...)
}

// List all users in the system.
func (m *multiClient) ListUsers(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListUsersResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListUsers(ctx, in, opts...)
}

// Update the details about an existing user.
func (m *multiClient) UpdateUser(ctx context.Context, id string, in *pb.UpdateUserRequest, opts ...grpc.CallOption) (*pb.UpdateUserResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpdateUser(ctx, in, opts...)
}

// Delete a user. This will invalidate all authentication for this user
// as well since they no longer exist.
func (m *multiClient) DeleteUser(ctx context.Context, id string, in *pb.DeleteUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.DeleteUser(ctx, in, opts...)
}

// UpsertAuthMethod upserts the auth method. All users logged in with
// this auth method will remain logged in even if settings change.
func (m *multiClient) UpsertAuthMethod(ctx context.Context, id string, in *pb.UpsertAuthMethodRequest, opts ...grpc.CallOption) (*pb.UpsertAuthMethodResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertAuthMethod(ctx, in, opts...)
}

// GetAuthMethod returns the auth method.
func (m *multiClient) GetAuthMethod(ctx context.Context, id string, in *pb.GetAuthMethodRequest, opts ...grpc.CallOption) (*pb.GetAuthMethodResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetAuthMethod(ctx, in, opts...)
}

// ListAuthMethods returns a list of all the auth methods.
func (m *multiClient) ListAuthMethods(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListAuthMethodsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListAuthMethods(ctx, in, opts...)
}

// Delete an auth method. This will invalidate all users authenticated
// using this auth method and they will have to reauthenticate some other
// way.
func (m *multiClient) DeleteAuthMethod(ctx context.Context, id string, in *pb.DeleteAuthMethodRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.DeleteAuthMethod(ctx, in, opts...)
}

// ListWorkspaces returns a list of all workspaces.
//
// Note that currently this list is never pruned, even if a workspace is
// no longer in use. We plan to prune this in a future improvement.
func (m *multiClient) ListWorkspaces(ctx context.Context, id string, in *pb.ListWorkspacesRequest, opts ...grpc.CallOption) (*pb.ListWorkspacesResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListWorkspaces(ctx, in, opts...)
}

// GetWorkspace returns the workspace.
func (m *multiClient) GetWorkspace(ctx context.Context, id string, in *pb.GetWorkspaceRequest, opts ...grpc.CallOption) (*pb.GetWorkspaceResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetWorkspace(ctx, in, opts...)
}

// UpsertWorkspace upserts the workspace. Changes to a Workspace's Projects
// are ignored at this time.
func (m *multiClient) UpsertWorkspace(ctx context.Context, id string, in *pb.UpsertWorkspaceRequest, opts ...grpc.CallOption) (*pb.UpsertWorkspaceResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertWorkspace(ctx, in, opts...)
}

// UpsertProject upserts the project.
func (m *multiClient) UpsertProject(ctx context.Context, id string, in *pb.UpsertProjectRequest, opts ...grpc.CallOption) (*pb.UpsertProjectResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertProject(ctx, in, opts...)
}

// GetProject returns the project.
func (m *multiClient) GetProject(ctx context.Context, id string, in *pb.GetProjectRequest, opts ...grpc.CallOption) (*pb.GetProjectResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetProject(ctx, in, opts...)
}

// ListProjects returns a list of all the projects. There is no equivalent
// ListApplications because applications are a part of projects and you
// can use GetProject to get more information about the project.
func (m *multiClient) ListProjects(ctx context.Context, id string, in *pb.ListProjectsRequest, opts ...grpc.CallOption) (*pb.ListProjectsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListProjects(ctx, in, opts...)
}

// DestroyProject deletes a project from the database as well as (optionally)
// destroys all resources created within a project
func (m *multiClient) DestroyProject(ctx context.Context, id string, in *pb.DestroyProjectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.DestroyProject(ctx, in, opts...)
}

// GetApplication returns one application on the project.
func (m *multiClient) GetApplication(ctx context.Context, id string, in *pb.GetApplicationRequest, opts ...grpc.CallOption) (*pb.GetApplicationResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetApplication(ctx, in, opts...)
}

// UpsertApplication upserts an application with a project.
func (m *multiClient) UpsertApplication(ctx context.Context, id string, in *pb.UpsertApplicationRequest, opts ...grpc.CallOption) (*pb.UpsertApplicationResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertApplication(ctx, in, opts...)
}

// ListBuilds returns the builds.
func (m *multiClient) ListBuilds(ctx context.Context, id string, in *pb.ListBuildsRequest, opts ...grpc.CallOption) (*pb.ListBuildsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListBuilds(ctx, in, opts...)
}

// GetBuild returns a build
func (m *multiClient) GetBuild(ctx context.Context, id string, in *pb.GetBuildRequest, opts ...grpc.CallOption) (*pb.Build, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetBuild(ctx, in, opts...)
}

// GetLatestBuild returns the most recent successfully completed build
// for an app.
func (m *multiClient) GetLatestBuild(ctx context.Context, id string, in *pb.GetLatestBuildRequest, opts ...grpc.CallOption) (*pb.Build, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetLatestBuild(ctx, in, opts...)
}

// ListPushedArtifacts returns the builds.
func (m *multiClient) ListPushedArtifacts(ctx context.Context, id string, in *pb.ListPushedArtifactsRequest, opts ...grpc.CallOption) (*pb.ListPushedArtifactsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListPushedArtifacts(ctx, in, opts...)
}

// GetPushedArtifact returns a deployment
func (m *multiClient) GetPushedArtifact(ctx context.Context, id string, in *pb.GetPushedArtifactRequest, opts ...grpc.CallOption) (*pb.PushedArtifact, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetPushedArtifact(ctx, in, opts...)
}

// GetLatestPushedArtifact returns the most recent successfully completed
// artifact push for an app.
func (m *multiClient) GetLatestPushedArtifact(ctx context.Context, id string, in *pb.GetLatestPushedArtifactRequest, opts ...grpc.CallOption) (*pb.PushedArtifact, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetLatestPushedArtifact(ctx, in, opts...)
}

// ListDeployments returns the deployments.
func (m *multiClient) ListDeployments(ctx context.Context, id string, in *pb.ListDeploymentsRequest, opts ...grpc.CallOption) (*pb.ListDeploymentsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListDeployments(ctx, in, opts...)
}

// GetDeployment returns a deployment
func (m *multiClient) GetDeployment(ctx context.Context, id string, in *pb.GetDeploymentRequest, opts ...grpc.CallOption) (*pb.Deployment, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetDeployment(ctx, in, opts...)
}

// ListInstances returns the running instances of deployments.
func (m *multiClient) ListInstances(ctx context.Context, id string, in *pb.ListInstancesRequest, opts ...grpc.CallOption) (*pb.ListInstancesResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListInstances(ctx, in, opts...)
}

// ListReleases returns the releases.
func (m *multiClient) ListReleases(ctx context.Context, id string, in *pb.ListReleasesRequest, opts ...grpc.CallOption) (*pb.ListReleasesResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListReleases(ctx, in, opts...)
}

// GetRelease returns a release
func (m *multiClient) GetRelease(ctx context.Context, id string, in *pb.GetReleaseRequest, opts ...grpc.CallOption) (*pb.Release, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetRelease(ctx, in, opts...)
}

// GetLatestRelease returns the most recent successfully completed
// release for an app.
func (m *multiClient) GetLatestRelease(ctx context.Context, id string, in *pb.GetLatestReleaseRequest, opts ...grpc.CallOption) (*pb.Release, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetLatestRelease(ctx, in, opts...)
}

// GetStatusReport returns a StatusReport
func (m *multiClient) GetStatusReport(ctx context.Context, id string, in *pb.GetStatusReportRequest, opts ...grpc.CallOption) (*pb.StatusReport, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetStatusReport(ctx, in, opts...)
}

// GetLatestStatusReport returns the most recent successfully completed
// health report for an app
func (m *multiClient) GetLatestStatusReport(ctx context.Context, id string, in *pb.GetLatestStatusReportRequest, opts ...grpc.CallOption) (*pb.StatusReport, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetLatestStatusReport(ctx, in, opts...)
}

// ListStatusReports returns the deployments.
func (m *multiClient) ListStatusReports(ctx context.Context, id string, in *pb.ListStatusReportsRequest, opts ...grpc.CallOption) (*pb.ListStatusReportsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListStatusReports(ctx, in, opts...)
}

// ExpediteStatusReport returns the queued status report job id
func (m *multiClient) ExpediteStatusReport(ctx context.Context, id string, in *pb.ExpediteStatusReportRequest, opts ...grpc.CallOption) (*pb.ExpediteStatusReportResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ExpediteStatusReport(ctx, in, opts...)
}

// GetLogStream reads the log stream for a deployment. This will immediately
// send a single LogEntry with the lines we have so far. If there are no
// available lines this will NOT block and instead will return an error.
// The client can choose to retry or not.
func (m *multiClient) GetLogStream(ctx context.Context, id string, in *pb.GetLogStreamRequest, opts ...grpc.CallOption) (pb.Waypoint_GetLogStreamClient, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetLogStream(ctx, in, opts...)
}

// StartExecStream starts an exec session.
func (m *multiClient) StartExecStream(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_StartExecStreamClient, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.StartExecStream(ctx, opts...)
}

// Set one or more configuration variables for applications or runners.
func (m *multiClient) SetConfig(ctx context.Context, id string, in *pb.ConfigSetRequest, opts ...grpc.CallOption) (*pb.ConfigSetResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.SetConfig(ctx, in, opts...)
}

// Retrieve merged configuration values for a specific scope. You can determine
// where a configuration variable was set by looking at the scope field on
// each variable.
func (m *multiClient) GetConfig(ctx context.Context, id string, in *pb.ConfigGetRequest, opts ...grpc.CallOption) (*pb.ConfigGetResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetConfig(ctx, in, opts...)
}

// Set the configuration for a dynamic configuration source. If you're looking
// to set application configuration, you probably want SetConfig instead.
func (m *multiClient) SetConfigSource(ctx context.Context, id string, in *pb.SetConfigSourceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.SetConfigSource(ctx, in, opts...)
}

// Get the matching configuration source for the request. This will return
// the most specific matching config source given the scope in the request.
// For example, if you search for an app-specific config source and only
// a global config exists, the global config will be returned.
func (m *multiClient) GetConfigSource(ctx context.Context, id string, in *pb.GetConfigSourceRequest, opts ...grpc.CallOption) (*pb.GetConfigSourceResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetConfigSource(ctx, in, opts...)
}

// Create a hostname with the URL service.
func (m *multiClient) CreateHostname(ctx context.Context, id string, in *pb.CreateHostnameRequest, opts ...grpc.CallOption) (*pb.CreateHostnameResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.CreateHostname(ctx, in, opts...)
}

// Delete a hostname with the URL service.
func (m *multiClient) DeleteHostname(ctx context.Context, id string, in *pb.DeleteHostnameRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.DeleteHostname(ctx, in, opts...)
}

// List all our registered hostnames.
func (m *multiClient) ListHostnames(ctx context.Context, id string, in *pb.ListHostnamesRequest, opts ...grpc.CallOption) (*pb.ListHostnamesResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListHostnames(ctx, in, opts...)
}

// QueueJob queues a job for execution by a runner. This will return as
// soon as the job is queued, it will not wait for execution.
func (m *multiClient) QueueJob(ctx context.Context, id string, in *pb.QueueJobRequest, opts ...grpc.CallOption) (*pb.QueueJobResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.QueueJob(ctx, in, opts...)
}

// CancelJob cancels a job. If the job is still queued this is a quick
// and easy operation. If the job is already completed, then this does
// nothing. If the job is assigned or running, then this will signal
// the runner about the cancellation but it may take time.
//
// This RPC always returns immediately. You must use GetJob or GetJobStream
// to wait on the status of the cancellation.
func (m *multiClient) CancelJob(ctx context.Context, id string, in *pb.CancelJobRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.CancelJob(ctx, in, opts...)
}

// GetJob queries a job by ID.
func (m *multiClient) GetJob(ctx context.Context, id string, in *pb.GetJobRequest, opts ...grpc.CallOption) (*pb.Job, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetJob(ctx, in, opts...)
}

// ListJobs will return a list of jobs known to Waypoint server. Can be filtered
// by request on values like workspace, project, application, job state, etc.
func (m *multiClient) ListJobs(ctx context.Context, id string, in *pb.ListJobsRequest, opts ...grpc.CallOption) (*pb.ListJobsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListJobs(ctx, in, opts...)
}

// ValidateJob checks if a job appears valid. This will check the job
// structure itself (i.e. missing fields) and can also check to ensure
// the job is assignable to a runner.
func (m *multiClient) ValidateJob(ctx context.Context, id string, in *pb.ValidateJobRequest, opts ...grpc.CallOption) (*pb.ValidateJobResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ValidateJob(ctx, in, opts...)
}

// GetJobStream opens a job event stream for a running job. This can be
// used to listen for terminal output and other events of a running job.
// Multiple listeners can
func (m *multiClient) GetJobStream(ctx context.Context, id string, in *pb.GetJobStreamRequest, opts ...grpc.CallOption) (pb.Waypoint_GetJobStreamClient, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetJobStream(ctx, in, opts...)
}

// GetRunner gets information about a single runner.
func (m *multiClient) GetRunner(ctx context.Context, id string, in *pb.GetRunnerRequest, opts ...grpc.CallOption) (*pb.Runner, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetRunner(ctx, in, opts...)
}

// ListRunners lists runners that are currently registered with the waypoint server.
// This list does not include previous on-demand runners that have exited.
func (m *multiClient) ListRunners(ctx context.Context, id string, in *pb.ListRunnersRequest, opts ...grpc.CallOption) (*pb.ListRunnersResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListRunners(ctx, in, opts...)
}

// AdoptRunners allows marking a runner as adopted or rejected.
func (m *multiClient) AdoptRunner(ctx context.Context, id string, in *pb.AdoptRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.AdoptRunner(ctx, in, opts...)
}

// ForgetRunner deletes an existing runner entry and makes the server
// behave as if the runner no longer exists. If the runner is currently
// running, it will receive errors on subsequent jobs, and will have to
// re-register. A forgotten runner will not be assigned new jobs until
// re-registered.
func (m *multiClient) ForgetRunner(ctx context.Context, id string, in *pb.ForgetRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ForgetRunner(ctx, in, opts...)
}

// GetServerConfig sets configuration for the Waypoint server.
func (m *multiClient) GetServerConfig(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.GetServerConfigResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetServerConfig(ctx, in, opts...)
}

// SetServerConfig sets configuration for the Waypoint server.
func (m *multiClient) SetServerConfig(ctx context.Context, id string, in *pb.SetServerConfigRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.SetServerConfig(ctx, in, opts...)
}

// CreateSnapshot creates a new database snapshot.
func (m *multiClient) CreateSnapshot(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (pb.Waypoint_CreateSnapshotClient, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.CreateSnapshot(ctx, in, opts...)
}

// RestoreSnapshot performs a database restore with the given snapshot.
// This API doesn't do a full online restore, it only stages the restore
// for the next server start to finalize the restore. See the arguments for
// more information.
func (m *multiClient) RestoreSnapshot(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_RestoreSnapshotClient, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.RestoreSnapshot(ctx, opts...)
}

// BootstrapToken returns the initial token for the server. This can only
// be requested once on first startup. After initial request this will
// always return a PermissionDenied error.
func (m *multiClient) BootstrapToken(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.NewTokenResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.BootstrapToken(ctx, in, opts...)
}

// DecodeToken takes a token string and returns the structured information
// about the given token. This is useful for frontends (CLI, UI, etc.) to
// learn more about a token before using it. For example, if a UI wants to
// create a signup flow around signup tokens, they can validate the token
// ahead of time.
//
// This endpoint does NOT require authentication.
func (m *multiClient) DecodeToken(ctx context.Context, id string, in *pb.DecodeTokenRequest, opts ...grpc.CallOption) (*pb.DecodeTokenResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.DecodeToken(ctx, in, opts...)
}

// Generate a new invite token that users can exchange for a login token.
// This can be used to also invite new users to the Waypoint server.
func (m *multiClient) GenerateInviteToken(ctx context.Context, id string, in *pb.InviteTokenRequest, opts ...grpc.CallOption) (*pb.NewTokenResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GenerateInviteToken(ctx, in, opts...)
}

// Generate a new login token that users can use to login directly.
// This can only be called for existing users.
func (m *multiClient) GenerateLoginToken(ctx context.Context, id string, in *pb.LoginTokenRequest, opts ...grpc.CallOption) (*pb.NewTokenResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GenerateLoginToken(ctx, in, opts...)
}

// Generate a new runner token that can be used with runners so they
// immediately begin work. The recommended appraoch is to instead use
// the adoption flow but this also works.
func (m *multiClient) GenerateRunnerToken(ctx context.Context, id string, in *pb.GenerateRunnerTokenRequest, opts ...grpc.CallOption) (*pb.NewTokenResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GenerateRunnerToken(ctx, in, opts...)
}

// Exchange a invite token for a login token. If the invite token is
// for a new user, this will create a new user account with the provided
// username hint.
func (m *multiClient) ConvertInviteToken(ctx context.Context, id string, in *pb.ConvertInviteTokenRequest, opts ...grpc.CallOption) (*pb.NewTokenResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ConvertInviteToken(ctx, in, opts...)
}

// RunnerToken is called to register a runner and request a token for
// remaining runner API calls. This kicks off the "adoption" process
// (if necessary).
//
// This is unauthenticated (but requires a cookie in the metadata).
func (m *multiClient) RunnerToken(ctx context.Context, id string, in *pb.RunnerTokenRequest, opts ...grpc.CallOption) (*pb.RunnerTokenResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.RunnerToken(ctx, in, opts...)
}

// RunnerConfig is called to receive the configuration for the runner.
// The response is a stream so that the configuration can be updated later.
func (m *multiClient) RunnerConfig(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_RunnerConfigClient, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.RunnerConfig(ctx, opts...)
}

// RunnerJobStream is called by a runner to request a single job for
// execution and update the status of that job.
func (m *multiClient) RunnerJobStream(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_RunnerJobStreamClient, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.RunnerJobStream(ctx, opts...)
}

// RunnerGetDeploymentConfig is called by a runner for a deployment operation
// to determine the settings to use for a deployment.
func (m *multiClient) RunnerGetDeploymentConfig(ctx context.Context, id string, in *pb.RunnerGetDeploymentConfigRequest, opts ...grpc.CallOption) (*pb.RunnerGetDeploymentConfigResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.RunnerGetDeploymentConfig(ctx, in, opts...)
}

// EntrypointConfig is called to get the configuration for the entrypoint
// and also to get any potential updates.
//
// This endpoint also registers the instance with the server. This MUST be
// called first otherwise other RPCs related to the entrypoint may fail
// with FailedPrecondition.
func (m *multiClient) EntrypointConfig(ctx context.Context, id string, in *pb.EntrypointConfigRequest, opts ...grpc.CallOption) (pb.Waypoint_EntrypointConfigClient, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.EntrypointConfig(ctx, in, opts...)
}

// EntrypointLogStream is called to open the stream that logs are sent to.
func (m *multiClient) EntrypointLogStream(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_EntrypointLogStreamClient, error) {
	panic("not implemented") // TODO: Implement
}

// EntrypointExecStream is called to open the data stream for the exec session.
func (m *multiClient) EntrypointExecStream(ctx context.Context, id string, opts ...grpc.CallOption) (pb.Waypoint_EntrypointExecStreamClient, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.EntrypointExecStream(ctx, opts...)
}

// WaypointHclFmt formats a waypoint.hcl file. This must be in HCL format.
// JSON formatting is not supported.
func (m *multiClient) WaypointHclFmt(ctx context.Context, id string, in *pb.WaypointHclFmtRequest, opts ...grpc.CallOption) (*pb.WaypointHclFmtResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.WaypointHclFmt(ctx, in, opts...)
}

// UpsertOnDemandRunnerConfig updates or inserts a on-demand runner
// configuration. This configuration can be used by projects for running
// operations on just-in-time launched runners.
func (m *multiClient) UpsertOnDemandRunnerConfig(ctx context.Context, id string, in *pb.UpsertOnDemandRunnerConfigRequest, opts ...grpc.CallOption) (*pb.UpsertOnDemandRunnerConfigResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertOnDemandRunnerConfig(ctx, in, opts...)
}

// GetOnDemandRunnerConfig returns the on-demand runner configuration.
func (m *multiClient) GetOnDemandRunnerConfig(ctx context.Context, id string, in *pb.GetOnDemandRunnerConfigRequest, opts ...grpc.CallOption) (*pb.GetOnDemandRunnerConfigResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetOnDemandRunnerConfig(ctx, in, opts...)
}

// GetOnDemandRunnerConfig returns the on-demand runner configuration.
func (m *multiClient) GetDefaultOnDemandRunnerConfig(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.GetOnDemandRunnerConfigResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetDefaultOnDemandRunnerConfig(ctx, in, opts...)
}

// GetOnDemandRunnerConfig returns the on-demand runner configuration.
func (m *multiClient) DeleteOnDemandRunnerConfig(ctx context.Context, id string, in *pb.DeleteOnDemandRunnerConfigRequest, opts ...grpc.CallOption) (*pb.DeleteOnDemandRunnerConfigResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.DeleteOnDemandRunnerConfig(ctx, in, opts...)
}

// ListOnDemandRunnerConfigs returns a list of all the on-demand runners configs.
func (m *multiClient) ListOnDemandRunnerConfigs(ctx context.Context, id string, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListOnDemandRunnerConfigsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListOnDemandRunnerConfigs(ctx, in, opts...)
}

// UpsertBuild updates or inserts a build. A build is responsible for
// taking some set of source information and turning it into an initial
// artifact. This artifact is considered "local" until it is pushed.
func (m *multiClient) UpsertBuild(ctx context.Context, id string, in *pb.UpsertBuildRequest, opts ...grpc.CallOption) (*pb.UpsertBuildResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertBuild(ctx, in, opts...)
}

// UpsertPushedArtifact updates or inserts a pushed artifact. This is
// useful for local operations to work on a pushed artifact.
func (m *multiClient) UpsertPushedArtifact(ctx context.Context, id string, in *pb.UpsertPushedArtifactRequest, opts ...grpc.CallOption) (*pb.UpsertPushedArtifactResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertPushedArtifact(ctx, in, opts...)
}

// UpsertDeployment updates or inserts a deployment.
func (m *multiClient) UpsertDeployment(ctx context.Context, id string, in *pb.UpsertDeploymentRequest, opts ...grpc.CallOption) (*pb.UpsertDeploymentResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertDeployment(ctx, in, opts...)
}

// UpsertRelease updates or inserts a release.
func (m *multiClient) UpsertRelease(ctx context.Context, id string, in *pb.UpsertReleaseRequest, opts ...grpc.CallOption) (*pb.UpsertReleaseResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertRelease(ctx, in, opts...)
}

// UpsertStatusReport updates or inserts a statusreport.
func (m *multiClient) UpsertStatusReport(ctx context.Context, id string, in *pb.UpsertStatusReportRequest, opts ...grpc.CallOption) (*pb.UpsertStatusReportResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertStatusReport(ctx, in, opts...)
}

// GetTask returns a requested Task message. Or an error if it does not exist.
func (m *multiClient) GetTask(ctx context.Context, id string, in *pb.GetTaskRequest, opts ...grpc.CallOption) (*pb.GetTaskResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetTask(ctx, in, opts...)
}

// ListTask will return a list of all existing Tasks
func (m *multiClient) ListTask(ctx context.Context, id string, in *pb.ListTaskRequest, opts ...grpc.CallOption) (*pb.ListTaskResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListTask(ctx, in, opts...)
}

// CancelTask will attempt to gracefully cancel each job in the task job triple
func (m *multiClient) CancelTask(ctx context.Context, id string, in *pb.CancelTaskRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.CancelTask(ctx, in, opts...)
}

// UpsertTrigger updates or inserts a trigger URL configuration.
func (m *multiClient) UpsertTrigger(ctx context.Context, id string, in *pb.UpsertTriggerRequest, opts ...grpc.CallOption) (*pb.UpsertTriggerResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertTrigger(ctx, in, opts...)
}

// GetTrigger returns a requested trigger message. Or an error if it does not exist.
func (m *multiClient) GetTrigger(ctx context.Context, id string, in *pb.GetTriggerRequest, opts ...grpc.CallOption) (*pb.GetTriggerResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetTrigger(ctx, in, opts...)
}

// DeleteTrigger takes a trigger id and deletes it, if it exists.
func (m *multiClient) DeleteTrigger(ctx context.Context, id string, in *pb.DeleteTriggerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.DeleteTrigger(ctx, in, opts...)
}

// ListTriggers takes a request filter, and returns any matching existing triggers
func (m *multiClient) ListTriggers(ctx context.Context, id string, in *pb.ListTriggerRequest, opts ...grpc.CallOption) (*pb.ListTriggerResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListTriggers(ctx, in, opts...)
}

// RunTrigger will look up the referenced trigger and attempt to queue a job
// based on the trigger configuration.
func (m *multiClient) RunTrigger(ctx context.Context, id string, in *pb.RunTriggerRequest, opts ...grpc.CallOption) (*pb.RunTriggerResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.RunTrigger(ctx, in, opts...)
}

// UpsertPipeline updates or inserts a pipeline. This is an INTERNAL ONLY
// endpoint that is meant to only be called by runners. Calling this manually
// can risk the internal state for pipelines. In the future, we'll restrict
// access to this via ACLs.
func (m *multiClient) UpsertPipeline(ctx context.Context, id string, in *pb.UpsertPipelineRequest, opts ...grpc.CallOption) (*pb.UpsertPipelineResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UpsertPipeline(ctx, in, opts...)
}

// RunPipeline queues a pipeline execution.
func (m *multiClient) RunPipeline(ctx context.Context, id string, in *pb.RunPipelineRequest, opts ...grpc.CallOption) (*pb.RunPipelineResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.RunPipeline(ctx, in, opts...)
}

// GetPipeline returns a pipeline proto by pipeline ref id
func (m *multiClient) GetPipeline(ctx context.Context, id string, in *pb.GetPipelineRequest, opts ...grpc.CallOption) (*pb.GetPipelineResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetPipeline(ctx, in, opts...)
}

// GetPipelineRun returns a pipeline run proto by pipeline ref id and sequence
func (m *multiClient) GetPipelineRun(ctx context.Context, id string, in *pb.GetPipelineRunRequest, opts ...grpc.CallOption) (*pb.GetPipelineRunResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetPipelineRun(ctx, in, opts...)
}

// GetLatestPipelineRun returns a pipeline run proto by pipeline ref id and sequence
func (m *multiClient) GetLatestPipelineRun(ctx context.Context, id string, in *pb.GetPipelineRequest, opts ...grpc.CallOption) (*pb.GetPipelineRunResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.GetLatestPipelineRun(ctx, in, opts...)
}

// ListPipelines takes a project and evaluates the projects config to get
// a list of Pipeline protos to return in the response. These pipelines
// are scoped to a single project from the request. It will return an
// error if the requested project does not exist, or an empty response
// if no pipelines are defined for the project.
func (m *multiClient) ListPipelines(ctx context.Context, id string, in *pb.ListPipelinesRequest, opts ...grpc.CallOption) (*pb.ListPipelinesResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListPipelines(ctx, in, opts...)
}

// ListPipelineRuns takes a pipeline ref and returns a list of runs of that pipeline.
// It will return an error if the requested pipeline does not exist, or an empty response
// if there are no runs for the pipeline.
func (m *multiClient) ListPipelineRuns(ctx context.Context, id string, in *pb.ListPipelineRunsRequest, opts ...grpc.CallOption) (*pb.ListPipelineRunsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ListPipelineRuns(ctx, in, opts...)
}

// ConfigSyncPipeline takes a request for a given project and syncs the current
// project config to the Waypoint database.
func (m *multiClient) ConfigSyncPipeline(ctx context.Context, id string, in *pb.ConfigSyncPipelineRequest, opts ...grpc.CallOption) (*pb.ConfigSyncPipelineResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.ConfigSyncPipeline(ctx, in, opts...)
}

// List full projects (not just refs)
func (m *multiClient) UI_ListProjects(ctx context.Context, id string, in *pb.UI_ListProjectsRequest, opts ...grpc.CallOption) (*pb.UI_ListProjectsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UI_ListProjects(ctx, in, opts...)
}

// Get a given project with useful related records.
func (m *multiClient) UI_GetProject(ctx context.Context, id string, in *pb.UI_GetProjectRequest, opts ...grpc.CallOption) (*pb.UI_GetProjectResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UI_GetProject(ctx, in, opts...)
}

// List deployments for a given application.
func (m *multiClient) UI_ListDeployments(ctx context.Context, id string, in *pb.UI_ListDeploymentsRequest, opts ...grpc.CallOption) (*pb.UI_ListDeploymentsResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UI_ListDeployments(ctx, in, opts...)
}

// GetDeployment returns a deployment
func (m *multiClient) UI_GetDeployment(ctx context.Context, id string, in *pb.UI_GetDeploymentRequest, opts ...grpc.CallOption) (*pb.UI_GetDeploymentResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UI_GetDeployment(ctx, in, opts...)
}

// List releases for a given application.
func (m *multiClient) UI_ListReleases(ctx context.Context, id string, in *pb.UI_ListReleasesRequest, opts ...grpc.CallOption) (*pb.UI_ListReleasesResponse, error) {
	c, err := m.getClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return c.UI_ListReleases(ctx, in, opts...)
}
