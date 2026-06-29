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
	// defaultRequestTimeout bounds every SDK call so a stalled HTTP/2
	// stream or a hung server response can never park a caller forever.
	// Callers that need shorter (heartbeats, polling) can pass a tighter
	// ctx; callers that need longer can override via WithRequestTimeout.
	defaultRequestTimeout = 60 * time.Second

	// HTTP/2 ping config — without these, a half-open stream on a
	// dropped LB connection blocks indefinitely even with
	// ResponseHeaderTimeout set on the HTTP/1 transport.
	defaultH2ReadIdleTimeout = 30 * time.Second
	defaultH2PingTimeout     = 15 * time.Second
)

// newDefaultTransport builds an *http.Transport that does not share state with
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

	// heartbeat and health checks
	CreateHeartBeat(ctx context.Context, req *models.ServiceCreateRunnerHeartBeatRequest) (*models.AppRunnerHeartBeat, error)
	CreateHealthCheck(ctx context.Context, req *models.ServiceCreateRunnerHealthCheckRequest) (*models.AppRunnerHealthCheck, error)

	// jobs
	GetJobs(ctx context.Context, grp models.AppRunnerJobGroup, status models.AppRunnerJobStatus, limit *int64) ([]*models.AppRunnerJob, error)
	TailJobs(ctx context.Context, grp models.AppRunnerJobGroup, wait time.Duration) ([]*models.AppRunnerJob, error)
	GetJob(ctx context.Context, jobID string) (*models.AppRunnerJob, error)
	GetJobPlanJSON(ctx context.Context, jobID string) (string, error)
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

	// get an app config
	GetAppConfig(ctx context.Context, appID, appConfigID string) (*models.AppAppConfig, error)

	// installs
	GetInstallComponenetLastActivePlan(ctx context.Context, installId, componentId string) (*models.ServiceGetInstallComponenetLastActivePlanResponse, error)

	UpdateTerraformStateJSON(ctx context.Context, workspaceID string, jobID *string, reqBody any) (any, error)

	LockTerraformWorkspace(ctx context.Context, workspaceID string, jobID *string, reqBody any) error
	UnlockTerraformWorkspace(ctx context.Context, workspaceID string) error

	// runner processes
	CreateProcess(ctx context.Context, req *models.ServiceCreateRunnerProcessRequest) (*models.AppRunnerProcess, error)
	GetProcess(ctx context.Context, processID string) (*models.AppRunnerProcess, error)
	GetProcessShutdowns(ctx context.Context, processID string) ([]*models.AppRunnerProcessShutdown, error)
	UpdateProcess(ctx context.Context, processID string, req *models.ServiceUpdateRunnerProcessRequest) (*models.AppRunnerProcess, error)
	CompleteShutdown(ctx context.Context, processID, shutdownID string) (*models.AppRunnerProcessShutdown, error)
	ReportTerminating(ctx context.Context, processID string) error

	// runner
	GetRunner(ctx context.Context) (*models.AppRunner, error)

	// sandbox configs
	GetSandboxConfigs(ctx context.Context) ([]*SandboxConfig, error)
	GetSandboxConfig(ctx context.Context, jobType, operation string) (*SandboxConfig, error)

	// authentication
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

	// http.Client.Timeout backstops every request — including a slow body
	// read, which ResponseHeaderTimeout does not cover. The caller's ctx
	// deadline still wins if shorter.
	c.httpClient = &http.Client{
		Transport: appTransport,
		Timeout:   c.RequestTimeout,
	}

	// unauthClient shares the tuned base transport (timeouts, HTTP/2 pings,
	// connection pool) but skips appTransport's Authorization injection. It is
	// used for endpoints that are public by design — runner-auth bootstrap and
	// shutdown polling — so SDK requests to those routes carry no auth header.
	c.unauthClient = &http.Client{
		Transport: base,
		Timeout:   c.RequestTimeout,
	}

	transport := httptransport.NewWithClient(apiURL.Host, apiURL.Path, []string{apiURL.Scheme}, c.httpClient)
	c.genClient = genclient.New(transport, nil)

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

// WithRequestTimeout overrides the default per-request timeout enforced by the
// underlying http.Client. A caller-supplied context deadline still wins if
// shorter. Pass 0 to rely solely on caller contexts (discouraged for loops).
func WithRequestTimeout(d time.Duration) clientOption {
	return func(c *client) error {
		c.RequestTimeout = d
		return nil
	}
}
