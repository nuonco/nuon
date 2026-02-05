package nuonrunner

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-playground/validator/v10"

	genclient "github.com/nuonco/nuon/sdks/nuon-runner-go/client"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

//go:generate ./generate.sh
type Client interface {
	SetRunnerID(runnerID string)

	GetSettings(ctx context.Context) (*models.AppRunnerGroupSettings, error)

	// heartbeat and health checks
	CreateHeartBeat(ctx context.Context, req *models.ServiceCreateRunnerHeartBeatRequest) (*models.AppRunnerHeartBeat, error)
	CreateHealthCheck(ctx context.Context, req *models.ServiceCreateRunnerHealthCheckRequest) (*models.AppRunnerHealthCheck, error)

	// jobs
	GetJobs(ctx context.Context, grp models.AppRunnerJobGroup, status models.AppRunnerJobStatus, limit *int64) ([]*models.AppRunnerJob, error)
	GetJob(ctx context.Context, jobID string) (*models.AppRunnerJob, error)
	GetJobPlanJSON(ctx context.Context, jobID string) (string, error) // DEPRECATED: Use GetJobCompositePlan instead
	GetJobCompositePlan(ctx context.Context, jobID string) (*models.PlantypesCompositePlan, error)
	UpdateJob(ctx context.Context, jobID string, req *models.ServiceUpdateRunnerJobRequest) (*models.AppRunnerJob, error)

	// job executions
	GetJobExecutions(ctx context.Context, jobID string) ([]*models.AppRunnerJobExecution, error)
	CreateJobExecution(ctx context.Context, jobID string, req *models.ServiceCreateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error)
	UpdateJobExecution(ctx context.Context, jobID, jobExecutionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error)
	CreateJobExecutionResult(ctx context.Context, jobID, jobExecutionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error)
	CreateJobExecutionOutputs(ctx context.Context, jobID, jobExecutionID string, req *models.ServiceCreateRunnerJobExecutionOutputsRequest) (*models.AppRunnerJobExecutionOutputs, error)

	// otel operations
	WriteOTELLogs(ctx context.Context, req interface{}) error
	WriteOTELTraces(ctx context.Context, req interface{}) error
	WriteOTELMetrics(ctx context.Context, req interface{}) error

	// actions specific endpoints
	UpdateInstallActionWorkflowRunStep(ctx context.Context, installID, workflowID, runID string, req *models.ServiceUpdateInstallActionWorkflowRunStepRequest) (*models.AppInstallActionWorkflowRunStep, error)
	GetInstallActionWorkflowRun(ctx context.Context, installID, runID string) (*models.AppInstallActionWorkflowRun, error)

	GetActionWorkflowConfig(ctx context.Context, workflowConfigID string) (*models.AppActionWorkflowConfig, error)
	GetActionWorkflowLatestConfig(ctx context.Context, workflowID string) (*models.AppActionWorkflowConfig, error)

	// get an app config
	GetAppConfig(ctx context.Context, appID, appConfigID string) (*models.AppAppConfig, error)

	// installs
	GetInstallComponenetLastActivePlan(ctx context.Context, installId, componentId string) (*models.ServiceGetInstallComponenetLastActivePlanResponse, error)

	UpdateTerraformStateJSON(ctx context.Context, workspaceID string, jobID *string, reqBody any) (any, error)

	LockTerraformWorkspace(ctx context.Context, workspaceID string, jobID *string, reqBody any) error
	UnlockTerraformWorkspace(ctx context.Context, workspaceID string) error

	// authentication
	RunnerAuthAWS(ctx context.Context, req *models.GithubComNuoncoNuonServicesCtlAPIInternalAppRunnerAuthServiceRunnerAuthAWSRequest) (*models.GithubComNuoncoNuonServicesCtlAPIInternalAppRunnerAuthServiceRunnerAuthAWSResponse, error)
}

var _ Client = (*client)(nil)

type client struct {
	v *validator.Validate

	APIURL   string `validate:"required"`
	APIToken string
	RunnerID string

	genClient    *genclient.NuonRunnerAPI
	appTransport *appTransport
	retryer      Retryer
}

type clientOption func(*client) error

func New(opts ...clientOption) (*client, error) {
	c := &client{
		retryer: &defaultRetryer{},
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.v == nil {
		c.v = validator.New()
	}

	if err := c.v.Struct(c); err != nil {
		return nil, err
	}

	apiURL, err := url.Parse(c.APIURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse api url: %w", err)
	}

	transport := httptransport.New(apiURL.Host, apiURL.Path, []string{apiURL.Scheme})
	appTransport := &appTransport{
		authToken: c.APIToken,
		transport: http.DefaultTransport,
	}
	c.appTransport = appTransport
	transport.Transport = appTransport
	genClient := genclient.New(transport, nil)
	c.genClient = genClient

	return c, nil
}

// WithAuthToken specifies the auth token to use
func WithAuthToken(token string) clientOption {
	return func(c *client) error {
		c.APIToken = token
		return nil
	}
}

// WithURL specifies the url to use
func WithURL(url string) clientOption {
	return func(c *client) error {
		c.APIURL = url
		return nil
	}
}

// WithRunnerID specifies the runner id to use
func WithRunnerID(runnerID string) clientOption {
	return func(c *client) error {
		c.RunnerID = runnerID
		return nil
	}
}

// WithValidator specifies a validator to use
func WithValidator(v *validator.Validate) clientOption {
	return func(c *client) error {
		c.v = v
		return nil
	}
}

// WithRetryer specifies a retryer to use
func WithRetryer(r Retryer) clientOption {
	return func(c *client) error {
		c.retryer = r
		return nil
	}
}
