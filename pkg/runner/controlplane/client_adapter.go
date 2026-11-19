package controlplane

import (
	"context"
	"fmt"
	"time"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/runner/errcapture"
)

// clientAdapter satisfies nuonrunner.Client for job handlers running in the
// control plane. install, when set (by NewInstallExecutor), is a real
// nuon-runner-go client built from the install's runner API URL, runner ID and
// a per-job token; it stands in for the methods the control-plane Client
// above can't serve from the database (app-config reads, Terraform workspace
// lock/unlock, state updates, ...), so those calls go through the runner API's
// existing org/app scoping unchanged. It is nil for the build-only executor,
// where those methods stay unsupported.
type clientAdapter struct {
	Client
	install nuonrunner.Client
}

var _ nuonrunner.Client = (*clientAdapter)(nil)

func (c *clientAdapter) SetRunnerID(string)  {}
func (c *clientAdapter) SetAuthToken(string) {}
func (c *clientAdapter) ForJob(jobID, token string) (nuonrunner.Client, error) {
	if c.install != nil {
		return c.install.ForJob(jobID, token)
	}
	return nil, unsupported("ForJob")
}
func (c *clientAdapter) ClaimJobToken(ctx context.Context, jobID string) (string, error) {
	if c.install != nil {
		return c.install.ClaimJobToken(ctx, jobID)
	}
	return "", unsupported("ClaimJobToken")
}

func (c *clientAdapter) CreateJobExecutionResult(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	if req != nil && !req.Success {
		if output := errcapture.Output(ctx); output != "" {
			if req.ErrorMetadata == nil {
				req.ErrorMetadata = make(map[string]string)
			}
			if req.ErrorMetadata[errcapture.MetadataKey] == "" {
				req.ErrorMetadata[errcapture.MetadataKey] = output
			}
		}
	}
	return c.Client.CreateJobExecutionResult(ctx, jobID, executionID, req)
}

func (c *clientAdapter) GetSettings(ctx context.Context) (*models.AppRunnerGroupSettings, error) {
	if c.install != nil {
		return c.install.GetSettings(ctx)
	}
	return nil, unsupported("GetSettings")
}

func (c *clientAdapter) CreateHeartBeat(context.Context, *models.ServiceCreateRunnerHeartBeatRequest) (*models.AppRunnerHeartBeat, error) {
	return nil, unsupported("CreateHeartBeat")
}

func (c *clientAdapter) CreateHealthCheck(context.Context, *models.ServiceCreateRunnerHealthCheckRequest) (*models.AppRunnerHealthCheck, error) {
	return nil, unsupported("CreateHealthCheck")
}
func (c *clientAdapter) CreateComponentHealth(context.Context, *models.ServiceCreateComponentHealthRequest) (*models.ServiceCreateComponentHealthResponse, error) {
	return nil, unsupported("CreateComponentHealth")
}
func (c *clientAdapter) GetRunnerInstallComponents(context.Context) (*models.ServiceRunnerInstallComponentsResponse, error) {
	return nil, unsupported("GetRunnerInstallComponents")
}
func (c *clientAdapter) PutComponentHealthContext(context.Context, string, []string, []string) error {
	return unsupported("PutComponentHealthContext")
}
func (c *clientAdapter) GetComponentHealthContext(context.Context) (string, []string, []string, error) {
	return "", nil, nil, unsupported("GetComponentHealthContext")
}
func (c *clientAdapter) GetJobs(context.Context, models.AppRunnerJobGroup, models.AppRunnerJobStatus, *int64) ([]*models.AppRunnerJob, error) {
	return nil, unsupported("GetJobs")
}

func (c *clientAdapter) TailJobs(context.Context, models.AppRunnerJobGroup, time.Duration) ([]*models.AppRunnerJob, error) {
	return nil, unsupported("TailJobs")
}
func (c *clientAdapter) GetJob(ctx context.Context, jobID string) (*models.AppRunnerJob, error) {
	return c.Client.GetJob(ctx, jobID)
}

func (c *clientAdapter) GetJobCompositePlan(ctx context.Context, jobID string) (*models.PlantypesCompositePlan, error) {
	return c.Client.GetJobCompositePlan(ctx, jobID)
}

func (c *clientAdapter) UpdateJob(context.Context, string, *models.ServiceUpdateRunnerJobRequest) (*models.AppRunnerJob, error) {
	return nil, unsupported("UpdateJob")
}

func (c *clientAdapter) GetJobExecutions(context.Context, string) ([]*models.AppRunnerJobExecution, error) {
	return nil, unsupported("GetJobExecutions")
}

func (c *clientAdapter) CreateJobExecution(context.Context, string, *models.ServiceCreateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error) {
	return nil, unsupported("CreateJobExecution")
}
func (c *clientAdapter) WriteOTELLogs(context.Context, interface{}) error    { return nil }
func (c *clientAdapter) WriteOTELTraces(context.Context, interface{}) error  { return nil }
func (c *clientAdapter) WriteOTELMetrics(context.Context, interface{}) error { return nil }
func (c *clientAdapter) UpdateInstallActionWorkflowRunStep(context.Context, string, string, string, *models.ServiceUpdateInstallActionWorkflowRunStepRequest) (*models.AppInstallActionWorkflowRunStep, error) {
	return nil, unsupported("UpdateInstallActionWorkflowRunStep")
}

func (c *clientAdapter) GetInstallActionWorkflowRun(context.Context, string, string) (*models.AppInstallActionWorkflowRun, error) {
	return nil, unsupported("GetInstallActionWorkflowRun")
}

func (c *clientAdapter) GetActionWorkflowConfig(context.Context, string) (*models.AppActionWorkflowConfig, error) {
	return nil, unsupported("GetActionWorkflowConfig")
}

func (c *clientAdapter) GetAppConfig(ctx context.Context, appID, branch string) (*models.AppAppConfig, error) {
	if c.install != nil {
		return c.install.GetAppConfig(ctx, appID, branch)
	}
	return nil, unsupported("GetAppConfig")
}

func (c *clientAdapter) GetInstallComponenetLastActivePlan(ctx context.Context, installID, componentID string) (*models.ServiceGetInstallComponenetLastActivePlanResponse, error) {
	if c.install != nil {
		return c.install.GetInstallComponenetLastActivePlan(ctx, installID, componentID)
	}
	return nil, unsupported("GetInstallComponenetLastActivePlan")
}

func (c *clientAdapter) UpdateTerraformStateJSON(ctx context.Context, workspaceID string, lockID *string, state any) (any, error) {
	if c.install != nil {
		return c.install.UpdateTerraformStateJSON(ctx, workspaceID, lockID, state)
	}
	return nil, unsupported("UpdateTerraformStateJSON")
}

func (c *clientAdapter) LockTerraformWorkspace(ctx context.Context, workspaceID string, lockID *string, lockInfo any) error {
	if c.install != nil {
		return c.install.LockTerraformWorkspace(ctx, workspaceID, lockID, lockInfo)
	}
	return unsupported("LockTerraformWorkspace")
}

func (c *clientAdapter) UnlockTerraformWorkspace(ctx context.Context, workspaceID string) error {
	if c.install != nil {
		return c.install.UnlockTerraformWorkspace(ctx, workspaceID)
	}
	return unsupported("UnlockTerraformWorkspace")
}

func (c *clientAdapter) CreateProcess(context.Context, *models.ServiceCreateRunnerProcessRequest) (*models.AppRunnerProcess, error) {
	return nil, unsupported("CreateProcess")
}

func (c *clientAdapter) GetProcess(context.Context, string) (*models.AppRunnerProcess, error) {
	return nil, unsupported("GetProcess")
}

func (c *clientAdapter) GetProcessShutdowns(context.Context, string) ([]*models.AppRunnerProcessShutdown, error) {
	return nil, unsupported("GetProcessShutdowns")
}

func (c *clientAdapter) UpdateProcess(context.Context, string, *models.ServiceUpdateRunnerProcessRequest) (*models.AppRunnerProcess, error) {
	return nil, unsupported("UpdateProcess")
}

func (c *clientAdapter) CompleteShutdown(context.Context, string, string) (*models.AppRunnerProcessShutdown, error) {
	return nil, unsupported("CompleteShutdown")
}

func (c *clientAdapter) ReportTerminating(context.Context, string) error {
	return unsupported("ReportTerminating")
}

func (c *clientAdapter) GetRunner(context.Context) (*models.AppRunner, error) {
	return nil, unsupported("GetRunner")
}

func (c *clientAdapter) CreateTelemetryAccessToken(context.Context, string) (*models.ServiceCreateTelemetryAccessTokenResponse, error) {
	return nil, unsupported("CreateTelemetryAccessToken")
}

func (c *clientAdapter) CreateInstallTelemetryAccessToken(context.Context, string, string) (*models.ServiceCreateInstallTelemetryAccessTokenResponse, error) {
	return nil, unsupported("CreateInstallTelemetryAccessToken")
}

func (c *clientAdapter) GetInstallTelemetryCollectorSettings(context.Context, string) (*models.ServiceInstallTelemetryCollectorSettings, error) {
	return nil, unsupported("GetInstallTelemetryCollectorSettings")
}

func (c *clientAdapter) GetSandboxConfigs(context.Context) ([]*nuonrunner.SandboxConfig, error) {
	return nil, unsupported("GetSandboxConfigs")
}

func (c *clientAdapter) GetSandboxConfig(context.Context, string, string) (*nuonrunner.SandboxConfig, error) {
	return nil, unsupported("GetSandboxConfig")
}

func (c *clientAdapter) RunnerAuthAWS(context.Context, *models.ServiceRunnerAuthAWSRequest) (*models.ServiceRunnerAuthAWSResponse, error) {
	return nil, unsupported("RunnerAuthAWS")
}

func (c *clientAdapter) RunnerAuthAWSIID(context.Context, *models.ServiceRunnerAuthAWSIIDRequest) (*models.ServiceRunnerAuthAWSIIDResponse, error) {
	return nil, unsupported("RunnerAuthAWSIID")
}

func (c *clientAdapter) RunnerAuthAWSSTS(context.Context, *models.ServiceRunnerAuthAWSSTSRequest) (*models.ServiceRunnerAuthAWSSTSResponse, error) {
	return nil, unsupported("RunnerAuthAWSSTS")
}

func (c *clientAdapter) RunnerAuthGCP(context.Context, *models.ServiceRunnerAuthGCPRequest) (*models.ServiceRunnerAuthGCPResponse, error) {
	return nil, unsupported("RunnerAuthGCP")
}

func (c *clientAdapter) RunnerAuthAzure(context.Context, *models.ServiceRunnerAuthAzureRequest) (*models.ServiceRunnerAuthAzureResponse, error) {
	return nil, unsupported("RunnerAuthAzure")
}

func (c *clientAdapter) GetPublicSettings(context.Context, string) (*models.ServiceRunnerPublicSettings, error) {
	return nil, unsupported("GetPublicSettings")
}

func unsupported(method string) error {
	return fmt.Errorf("control-plane runner client does not support %s", method)
}
