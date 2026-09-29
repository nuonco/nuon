package nuonrunner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-playground/validator/v10"
	"golang.org/x/net/http2"

	genclient "github.com/nuonco/nuon/sdks/nuon-runner-go/client"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	// why: defaultRequestTimeout bounds every SDK call so a stalled HTTP/2
	// stream or a hung server response can never park a caller forever.
	// Callers that need shorter (heartbeats, polling) can pass a tighter
	// ctx; callers that need longer can override via WithRequestTimeout.
	defaultRequestTimeout = 60 * time.Second

	defaultH2ReadIdleTimeout = 30 * time.Second
	defaultH2PingTimeout     = 15 * time.Second
)

// why: newDefaultTransport builds an *http.Transport that does not share state with
// http.DefaultTransport. Sharing the default is unsafe: other code in the
// runner mutates http.DefaultTransport globally (see helm chart packaging),
// which silently invalidates our connection pool and obscures network errors.
func newDefaultTransport() *http.Transport {
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          16,
		MaxIdleConnsPerHost:   8,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if h2, err := http2.ConfigureTransports(t); err == nil {
		h2.ReadIdleTimeout = defaultH2ReadIdleTimeout
		h2.PingTimeout = defaultH2PingTimeout
	}

	return t
}

//go:generate ./generate.sh
type Client interface {
	SetRunnerID(runnerID string)
	SetAuthToken(token string)

	GetSettings(ctx context.Context) (*models.AppRunnerGroupSettings, error)

	CreateHeartBeat(ctx context.Context, req *models.ServiceCreateRunnerHeartBeatRequest) (*models.AppRunnerHeartBeat, error)
	CreateHealthCheck(ctx context.Context, req *models.ServiceCreateRunnerHealthCheckRequest) (*models.AppRunnerHealthCheck, error)
	CreateComponentHealth(ctx context.Context, req *models.ServiceCreateComponentHealthRequest) (*models.ServiceCreateComponentHealthResponse, error)
	GetRunnerInstallComponents(ctx context.Context) (*models.ServiceRunnerInstallComponentsResponse, error)
	PutComponentHealthContext(ctx context.Context, clusterInfoJSON string, sandboxReleases, componentKinds []string) error
	GetComponentHealthContext(ctx context.Context) (string, []string, []string, error)

	GetJobs(ctx context.Context, grp models.AppRunnerJobGroup, status models.AppRunnerJobStatus, limit *int64) ([]*models.AppRunnerJob, error)
	TailJobs(ctx context.Context, grp models.AppRunnerJobGroup, wait time.Duration) ([]*models.AppRunnerJob, error)
	GetJob(ctx context.Context, jobID string) (*models.AppRunnerJob, error)
	// Deprecated: use GetJobCompositePlan.
	GetJobPlanJSON(ctx context.Context, jobID string) (string, error)
	GetJobCompositePlan(ctx context.Context, jobID string) (*models.PlantypesCompositePlan, error)
	UpdateJob(ctx context.Context, jobID string, req *models.ServiceUpdateRunnerJobRequest) (*models.AppRunnerJob, error)

	GetJobExecutions(ctx context.Context, jobID string) ([]*models.AppRunnerJobExecution, error)
	CreateJobExecution(ctx context.Context, jobID string, req *models.ServiceCreateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error)
	UpdateJobExecution(ctx context.Context, jobID, jobExecutionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error)
	CreateJobExecutionResult(ctx context.Context, jobID, jobExecutionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error)
	CreateJobExecutionOutputs(ctx context.Context, jobID, jobExecutionID string, req *models.ServiceCreateRunnerJobExecutionOutputsRequest) (*models.AppRunnerJobExecutionOutputs, error)

	WriteOTELLogs(ctx context.Context, req interface{}) error
	WriteOTELTraces(ctx context.Context, req interface{}) error
	WriteOTELMetrics(ctx context.Context, req interface{}) error

	UpdateInstallActionWorkflowRunStep(ctx context.Context, installID, workflowID, runID string, req *models.ServiceUpdateInstallActionWorkflowRunStepRequest) (*models.AppInstallActionWorkflowRunStep, error)
	GetInstallActionWorkflowRun(ctx context.Context, installID, runID string) (*models.AppInstallActionWorkflowRun, error)

	GetActionWorkflowConfig(ctx context.Context, workflowConfigID string) (*models.AppActionWorkflowConfig, error)

	GetAppConfig(ctx context.Context, appID, appConfigID string) (*models.AppAppConfig, error)

	GetInstallComponenetLastActivePlan(ctx context.Context, installId, componentId string) (*models.ServiceGetInstallComponenetLastActivePlanResponse, error)

	UpdateTerraformStateJSON(ctx context.Context, workspaceID string, jobID *string, reqBody any) (any, error)

	LockTerraformWorkspace(ctx context.Context, workspaceID string, jobID *string, reqBody any) error
	UnlockTerraformWorkspace(ctx context.Context, workspaceID string) error

	CreateProcess(ctx context.Context, req *models.ServiceCreateRunnerProcessRequest) (*models.AppRunnerProcess, error)
	GetProcess(ctx context.Context, processID string) (*models.AppRunnerProcess, error)
	GetProcessShutdowns(ctx context.Context, processID string) ([]*models.AppRunnerProcessShutdown, error)
	UpdateProcess(ctx context.Context, processID string, req *models.ServiceUpdateRunnerProcessRequest) (*models.AppRunnerProcess, error)
	CompleteShutdown(ctx context.Context, processID, shutdownID string) (*models.AppRunnerProcessShutdown, error)
	ReportTerminating(ctx context.Context, processID string) error

	GetRunner(ctx context.Context) (*models.AppRunner, error)
	CreateTelemetryAccessToken(ctx context.Context) (*models.ServiceCreateTelemetryAccessTokenResponse, error)

	GetSandboxConfigs(ctx context.Context) ([]*SandboxConfig, error)
	GetSandboxConfig(ctx context.Context, jobType, operation string) (*SandboxConfig, error)

	RunnerAuthAWS(ctx context.Context, req *models.ServiceRunnerAuthAWSRequest) (*models.ServiceRunnerAuthAWSResponse, error)
	RunnerAuthAWSIID(ctx context.Context, req *models.ServiceRunnerAuthAWSIIDRequest) (*models.ServiceRunnerAuthAWSIIDResponse, error)
	RunnerAuthGCP(ctx context.Context, req *models.ServiceRunnerAuthGCPRequest) (*models.ServiceRunnerAuthGCPResponse, error)
	RunnerAuthAzure(ctx context.Context, req *models.ServiceRunnerAuthAzureRequest) (*models.ServiceRunnerAuthAzureResponse, error)
}

var _ Client = (*client)(nil)

type client struct {
	v *validator.Validate

	APIURL         string `validate:"required"`
	APIToken       string
	RunnerID       string
	RequestTimeout time.Duration

	genClient    *genclient.NuonRunnerAPI
	appTransport *appTransport
	httpClient   *http.Client
	unauthClient *http.Client
	retryer      Retryer
}

type clientOption func(*client) error

func New(opts ...clientOption) (*client, error) {
	c := &client{
		retryer:        &defaultRetryer{},
		RequestTimeout: defaultRequestTimeout,
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

	base := newDefaultTransport()
	appTransport := &appTransport{
		authToken: c.APIToken,
		transport: base,
	}
	c.appTransport = appTransport

	c.httpClient = &http.Client{
		Transport: appTransport,
		Timeout:   c.RequestTimeout,
	}

	c.unauthClient = &http.Client{
		Transport: base,
		Timeout:   c.RequestTimeout,
	}

	transport := httptransport.NewWithClient(apiURL.Host, apiURL.Path, []string{apiURL.Scheme}, c.httpClient)
	c.genClient = genclient.New(transport, nil)

	return c, nil
}

func WithAuthToken(token string) clientOption {
	return func(c *client) error {
		c.APIToken = token
		return nil
	}
}

func WithURL(url string) clientOption {
	return func(c *client) error {
		c.APIURL = url
		return nil
	}
}

func WithRunnerID(runnerID string) clientOption {
	return func(c *client) error {
		c.RunnerID = runnerID
		return nil
	}
}

func WithValidator(v *validator.Validate) clientOption {
	return func(c *client) error {
		c.v = v
		return nil
	}
}

func WithRetryer(r Retryer) clientOption {
	return func(c *client) error {
		c.retryer = r
		return nil
	}
}

func WithRequestTimeout(d time.Duration) clientOption {
	return func(c *client) error {
		c.RequestTimeout = d
		return nil
	}
}
