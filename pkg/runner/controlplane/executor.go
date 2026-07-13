package controlplane

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	runnerctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	containerimagebuild "github.com/nuonco/nuon/pkg/runner/jobs/build/containerimage"
	helmbuild "github.com/nuonco/nuon/pkg/runner/jobs/build/helm"
	kubernetesmanifestbuild "github.com/nuonco/nuon/pkg/runner/jobs/build/kubernetes_manifest"
	noopbuild "github.com/nuonco/nuon/pkg/runner/jobs/build/noop"
	pulumibuild "github.com/nuonco/nuon/pkg/runner/jobs/build/pulumi"
	sandboxbuild "github.com/nuonco/nuon/pkg/runner/jobs/build/sandbox"
	terraformbuild "github.com/nuonco/nuon/pkg/runner/jobs/build/terraform"
	imagemetadatasync "github.com/nuonco/nuon/pkg/runner/jobs/sync/imagemetadata"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
	ociresolve "github.com/nuonco/nuon/pkg/runner/oci/resolve"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

type Client interface {
	GetJobPlanJSON(ctx context.Context, jobID string) (string, error)
	GetJobCompositePlan(ctx context.Context, jobID string) (*models.PlantypesCompositePlan, error)
	UpdateJobExecution(ctx context.Context, jobID, executionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error)
	CreateJobExecutionResult(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error)
	CreateJobExecutionOutputs(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionOutputsRequest) (*models.AppRunnerJobExecutionOutputs, error)
	WriteControlPlaneLogs(ctx context.Context, logStreamID string, records []OTELLogRecord) error
	WriteControlPlaneTraces(ctx context.Context, runnerID string, records []OTELTraceRecord) error
}

type Config struct {
	GitRef                   string
	BundleDir                string
	RegistryDir              string
	RegistryPort             int
	LogLevel                 string
	TerraformMirrorPlatforms []string
}

type Executor struct {
	client   Client
	l        *zap.Logger
	handlers []jobs.JobHandler
}

func NewExecutor(client Client, l *zap.Logger, cfg Config) (*Executor, error) {
	if l == nil {
		l = zap.NewNop()
	}
	if cfg.GitRef == "" {
		cfg.GitRef = "dev"
	}
	if cfg.BundleDir == "" {
		cfg.BundleDir = "/bundle"
	}
	if cfg.RegistryDir == "" {
		cfg.RegistryDir = "/tmp/runner-registry"
	}
	if cfg.RegistryPort == 0 {
		cfg.RegistryPort = 5001
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "INFO"
	}

	runnerCfg := &runnerconfig.Config{
		GitRef:                   cfg.GitRef,
		RunnerAPIURL:             "control-plane",
		RunnerID:                 "control-plane",
		HostIP:                   "127.0.0.1",
		LogLevel:                 cfg.LogLevel,
		BundleDir:                cfg.BundleDir,
		RegistryDir:              cfg.RegistryDir,
		RegistryPort:             cfg.RegistryPort,
		TerraformMirrorPlatforms: cfg.TerraformMirrorPlatforms,
	}
	v := validator.New()
	recorder := errs.NewRecorder(errs.Params{L: l, Settings: &settings.Settings{}, LC: noopLifecycle{}})
	copier := ocicopy.New(ocicopy.CopierParams{V: v, Cfg: runnerCfg})
	resolver := ociresolve.New(ociresolve.ResolverParams{V: v, Cfg: runnerCfg})
	apiClient := &clientAdapter{Client: client}

	handlers := make([]jobs.JobHandler, 0, 8)
	add := func(handler jobs.JobHandler, err error) error {
		if err != nil {
			return err
		}
		handlers = append(handlers, handler)
		return nil
	}
	if err := add(containerimagebuild.New(containerimagebuild.HandlerParams{V: v, APIClient: apiClient, Config: runnerCfg, OCICopy: copier, OCIResolve: resolver, ErrRecorder: recorder})); err != nil {
		return nil, err
	}
	if err := add(helmbuild.New(helmbuild.HandlerParams{V: v, APIClient: apiClient, Config: runnerCfg, OCICopy: copier, ErrRecorder: recorder})); err != nil {
		return nil, err
	}
	if err := add(kubernetesmanifestbuild.New(kubernetesmanifestbuild.HandlerParams{V: v, APIClient: apiClient, Config: runnerCfg, OCICopy: copier, ErrRecorder: recorder})); err != nil {
		return nil, err
	}
	if err := add(terraformbuild.New(terraformbuild.Params{V: v, APIClient: apiClient, Config: runnerCfg, OCICopy: copier, ErrRecorder: recorder})); err != nil {
		return nil, err
	}
	if err := add(pulumibuild.New(pulumibuild.Params{V: v, APIClient: apiClient, Config: runnerCfg, OCICopy: copier, ErrRecorder: recorder})); err != nil {
		return nil, err
	}
	if err := add(noopbuild.New(noopbuild.HandlerParams{V: v, APIClient: apiClient, Config: runnerCfg, ErrRecorder: recorder})); err != nil {
		return nil, err
	}
	if err := add(sandboxbuild.New(sandboxbuild.HandlerParams{V: v, APIClient: apiClient, Config: runnerCfg, OCICopy: copier, ErrRecorder: recorder})); err != nil {
		return nil, err
	}
	if err := add(imagemetadatasync.New(imagemetadatasync.HandlerParams{V: v, APIClient: apiClient, Config: runnerCfg, ErrRecorder: recorder})); err != nil {
		return nil, err
	}

	return &Executor{client: client, l: l, handlers: handlers}, nil
}

func (e *Executor) Execute(ctx context.Context, job *models.AppRunnerJob, execution *models.AppRunnerJobExecution) error {
	l := e.l
	logProvider, traceProvider := e.telemetryProviders(job)
	if logProvider != nil {
		l = newJobLogger(logProvider, e.l)
	}
	defer flushTelemetry(logProvider, traceProvider, l)

	l = l.With(
		zap.String("runner_job.id", job.ID),
		zap.String("runner_job.type", string(job.Type)),
		zap.String("runner_job.executor", "control-plane"),
		zap.String("runner_job_execution.id", execution.ID),
		zap.String("log_stream.id", job.LogStreamID),
	)

	handler, err := e.handler(job)
	if err != nil {
		l.Error("unable to resolve job handler", zap.Error(err))
		return err
	}

	ctx = runnerctx.SetJobMetadata(ctx, runnerctx.JobMetadata{RunnerJobID: job.ID, RunnerJobExecutionID: execution.ID})
	var tracer trace.Tracer
	var rootSpan trace.Span
	var jobErr error
	if traceProvider != nil {
		ctx = runnerctx.SetTracerProvider(ctx, traceProvider)
		tracer = traceProvider.Tracer("github.com/nuonco/nuon/pkg/runner/controlplane")
		ctx, rootSpan = tracer.Start(ctx, "job."+string(job.Type),
			trace.WithSpanKind(trace.SpanKindInternal),
			trace.WithAttributes(
				attribute.String("nuon.tool", "runner"),
				attribute.String("nuon.job.type", string(job.Type)),
				attribute.String("nuon.job.operation", string(job.Operation)),
				attribute.String("runner_job.id", job.ID),
				attribute.String("runner_job.executor", "control-plane"),
				attribute.String("runner_job_execution.id", execution.ID),
			),
		)
		l = l.With(runnerctx.ContextField(ctx))
		defer func() {
			if jobErr != nil {
				rootSpan.RecordError(jobErr)
				rootSpan.SetStatus(codes.Error, jobErr.Error())
			}
			rootSpan.End()
		}()
	}
	ctx = runnerctx.SetLogger(ctx, l)

	if isSandboxableBuildHandler(handler) {
		sandboxMode, err := e.getSandboxMode(ctx, job.ID)
		if err != nil {
			jobErr = fmt.Errorf("unable to check sandbox mode: %w", err)
			_, _ = e.client.UpdateJobExecution(ctx, job.ID, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
				Status:            statusForError(jobErr),
				StatusDescription: jobErr.Error(),
			})
			return jobErr
		}
		if sandboxMode != nil {
			l.Info("sandbox mode active for control-plane build; skipping real build execution")
			if err := e.executeSandboxBuild(ctx, job, execution, sandboxMode); err != nil {
				jobErr = fmt.Errorf("sandbox build: %w", err)
				_, _ = e.client.UpdateJobExecution(ctx, job.ID, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
					Status:            statusForError(jobErr),
					StatusDescription: jobErr.Error(),
				})
				return jobErr
			}
			return nil
		}
	}

	steps := []struct {
		name   string
		status models.AppRunnerJobExecutionStatus
		fn     func(context.Context) error
	}{
		{name: "resetting", status: models.AppRunnerJobExecutionStatusInitializing, fn: func(ctx context.Context) error {
			if h, ok := handler.(jobs.StatefulJobHandler); ok {
				return h.Reset(ctx)
			}
			return nil
		}},
		{name: "fetching", status: models.AppRunnerJobExecutionStatusInitializing, fn: func(ctx context.Context) error { return handler.Fetch(ctx, job, execution) }},
		{name: "validate", status: models.AppRunnerJobExecutionStatusInitializing, fn: func(ctx context.Context) error { return handler.Validate(ctx, job, execution) }},
		{name: "initialize", status: models.AppRunnerJobExecutionStatusInitializing, fn: func(ctx context.Context) error { return handler.Initialize(ctx, job, execution) }},
		{name: "execute", status: models.AppRunnerJobExecutionStatusInDashProgress, fn: func(ctx context.Context) error { return handler.Exec(ctx, job, execution) }},
		{name: "outputs", status: models.AppRunnerJobExecutionStatusInDashProgress, fn: func(ctx context.Context) error {
			outputs, err := handler.Outputs(ctx)
			if err != nil {
				return err
			}
			_, err = e.client.CreateJobExecutionOutputs(ctx, job.ID, execution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{Outputs: outputs})
			return err
		}},
		{name: "cleanup", status: models.AppRunnerJobExecutionStatusCleaningDashUp, fn: func(ctx context.Context) error { return handler.Cleanup(ctx, job, execution) }},
	}

	for _, step := range steps {
		if _, err := e.client.UpdateJobExecution(ctx, job.ID, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: step.status}); err != nil {
			jobErr = fmt.Errorf("unable to mark step %s started: %w", step.name, err)
			return jobErr
		}
		stepCtx := runnerctx.SetJobMetadata(ctx, runnerctx.JobMetadata{RunnerJobID: job.ID, RunnerJobExecutionID: execution.ID, StepName: step.name})
		stepL := l.With(zap.String("runner_job_execution_step.name", step.name))
		var stepSpan trace.Span
		if tracer != nil {
			stepCtx, stepSpan = tracer.Start(stepCtx, "step."+step.name,
				trace.WithSpanKind(trace.SpanKindInternal),
				trace.WithAttributes(
					attribute.String("nuon.tool", "runner"),
					attribute.String("runner_job_execution_step.name", step.name),
					attribute.String("runner_job.id", job.ID),
					attribute.String("runner_job.executor", "control-plane"),
					attribute.String("runner_job_execution.id", execution.ID),
				),
			)
			stepL = stepL.With(runnerctx.ContextField(stepCtx))
		}
		stepCtx = runnerctx.SetLogger(stepCtx, stepL)
		stepL.Info("executing job step "+step.name, zap.String("step", step.name))
		if err := step.fn(stepCtx); err != nil {
			stepL.Error("job step "+step.name+" failed", zap.String("step", step.name), zap.Error(err))
			if stepSpan != nil {
				stepSpan.RecordError(err)
				stepSpan.SetStatus(codes.Error, err.Error())
				stepSpan.End()
			}
			_, _ = e.client.UpdateJobExecution(ctx, job.ID, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
				Status:            statusForError(err),
				StatusDescription: fmt.Sprintf("%s: %s", step.name, err.Error()),
			})
			if step.name == "execute" || step.name == "outputs" {
				_ = handler.Cleanup(ctx, job, execution)
			}
			jobErr = fmt.Errorf("%s: %w", step.name, err)
			return jobErr
		}
		if stepSpan != nil {
			stepSpan.End()
		}
	}

	_, err = e.client.UpdateJobExecution(ctx, job.ID, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: models.AppRunnerJobExecutionStatusFinished})
	if err != nil {
		jobErr = err
	}
	return err
}

func (e *Executor) handler(job *models.AppRunnerJob) (jobs.JobHandler, error) {
	if job.Type == models.AppRunnerJobTypeDockerDashBuild {
		return nil, fmt.Errorf("docker_build components are not supported by control-plane builds; replace this component with a container_image component")
	}

	job.Status = models.AppRunnerJobStatusAvailable
	for _, handler := range e.handlers {
		if err := jobs.Matches(job, handler); err == nil {
			return handler, nil
		}
	}
	return nil, fmt.Errorf("job handler not found for %s job", job.Type)
}

func statusForError(err error) models.AppRunnerJobExecutionStatus {
	if errors.Is(err, context.DeadlineExceeded) {
		return models.AppRunnerJobExecutionStatusTimedDashOut
	}
	return models.AppRunnerJobExecutionStatusFailed
}

type noopLifecycle struct{}

func (noopLifecycle) Append(fx.Hook) {}

func HeartbeatUntilDone(ctx context.Context, heartbeat func()) func() {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-ticker.C:
				heartbeat()
			}
		}
	}()
	return func() { close(done) }
}
