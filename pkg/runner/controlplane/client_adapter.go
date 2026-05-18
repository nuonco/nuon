package controlplane

import (
	"context"
	"fmt"
	"time"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type clientAdapter struct {
	Client
}

var _ nuonrunner.Client = (*clientAdapter)(nil)

func (c *clientAdapter) SetRunnerID(string)  {}
func (c *clientAdapter) SetAuthToken(string) {}

func (c *clientAdapter) GetSettings(context.Context) (*models.AppRunnerGroupSettings, error) {
	return nil, unsupported("GetSettings")
}
func (c *clientAdapter) CreateHeartBeat(context.Context, *models.ServiceCreateRunnerHeartBeatRequest) (*models.AppRunnerHeartBeat, error) {
	return nil, unsupported("CreateHeartBeat")
}
func (c *clientAdapter) CreateHealthCheck(context.Context, *models.ServiceCreateRunnerHealthCheckRequest) (*models.AppRunnerHealthCheck, error) {
	return nil, unsupported("CreateHealthCheck")
}
func (c *clientAdapter) GetJobs(context.Context, models.AppRunnerJobGroup, models.AppRunnerJobStatus, *int64) ([]*models.AppRunnerJob, error) {
	return nil, unsupported("GetJobs")
}
func (c *clientAdapter) TailJobs(context.Context, models.AppRunnerJobGroup, time.Duration) ([]*models.AppRunnerJob, error) {
	return nil, unsupported("TailJobs")
}
func (c *clientAdapter) GetJob(context.Context, string) (*models.AppRunnerJob, error) {
	return nil, unsupported("GetJob")
}
func (c *clientAdapter) GetJobCompositePlan(context.Context, string) (*models.PlantypesCompositePlan, error) {
	return nil, unsupported("GetJobCompositePlan")
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
func (c *clientAdapter) GetAppConfig(context.Context, string, string) (*models.AppAppConfig, error) {
	return nil, unsupported("GetAppConfig")
}
func (c *clientAdapter) GetInstallComponenetLastActivePlan(context.Context, string, string) (*models.ServiceGetInstallComponenetLastActivePlanResponse, error) {
	return nil, unsupported("GetInstallComponenetLastActivePlan")
}
func (c *clientAdapter) UpdateTerraformStateJSON(context.Context, string, *string, any) (any, error) {
	return nil, unsupported("UpdateTerraformStateJSON")
}
func (c *clientAdapter) LockTerraformWorkspace(context.Context, string, *string, any) error {
	return unsupported("LockTerraformWorkspace")
}
func (c *clientAdapter) UnlockTerraformWorkspace(context.Context, string) error {
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
func (c *clientAdapter) RunnerAuthGCP(context.Context, *models.ServiceRunnerAuthGCPRequest) (*models.ServiceRunnerAuthGCPResponse, error) {
	return nil, unsupported("RunnerAuthGCP")
}
func (c *clientAdapter) RunnerAuthAzure(context.Context, *models.ServiceRunnerAuthAzureRequest) (*models.ServiceRunnerAuthAzureResponse, error) {
	return nil, unsupported("RunnerAuthAzure")
}

func unsupported(method string) error {
	return fmt.Errorf("control-plane runner client does not support %s", method)
}
