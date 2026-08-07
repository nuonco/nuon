package jobloop

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type jobLoopTestClient struct {
	nuonrunner.Client
	createResult    func(context.Context, string, string, *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error)
	updateExecution func(context.Context, string, string, *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error)
}

func (c *jobLoopTestClient) CreateJobExecutionResult(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	return c.createResult(ctx, jobID, executionID, req)
}

func (c *jobLoopTestClient) UpdateJobExecution(ctx context.Context, jobID, executionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error) {
	return c.updateExecution(ctx, jobID, executionID, req)
}

type jobLoopTestMetrics struct {
	metrics.Writer
}

func (*jobLoopTestMetrics) Incr(string, []string)                  {}
func (*jobLoopTestMetrics) Timing(string, time.Duration, []string) {}

type jobLoopTestHandler struct {
	jobs.JobHandler
	name string
}

func (h *jobLoopTestHandler) Name() string { return h.name }

func TestExecJobStepWritesFallbackResultBeforeFailedStatus(t *testing.T) {
	events := make([]string, 0, 3)
	var resultReq *models.ServiceCreateRunnerJobExecutionResultRequest
	client := &jobLoopTestClient{
		createResult: func(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
			events = append(events, "result")
			resultReq = req
			return &models.AppRunnerJobExecutionResult{ID: "result-id"}, nil
		},
		updateExecution: func(ctx context.Context, jobID, executionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error) {
			events = append(events, "status:"+string(req.Status))
			return &models.AppRunnerJobExecution{ID: executionID}, nil
		},
	}
	j := &jobLoop{apiClient: client, mw: &jobLoopTestMetrics{}}
	handlerErr := errors.New("terraform apply failed")
	step := &executeJobStep{
		name:        "execute",
		handler:     &jobLoopTestHandler{name: "terraform"},
		startStatus: models.AppRunnerJobExecutionStatusInDashProgress,
		fn: func(context.Context, jobs.JobHandler, *models.AppRunnerJob, *models.AppRunnerJobExecution) error {
			return handlerErr
		},
	}
	job := &models.AppRunnerJob{ID: "job-id", Type: models.AppRunnerJobTypeTerraformDashDeploy}
	execution := &models.AppRunnerJobExecution{ID: "execution-id"}

	err := j.execJobStep(context.Background(), zap.NewNop(), nil, step, job, execution)
	require.ErrorIs(t, err, handlerErr)
	require.Equal(t, []string{
		"status:in-progress",
		"result",
		"status:failed",
	}, events)
	require.False(t, resultReq.Success)
	require.Equal(t, map[string]string{
		"step":     "execute",
		"handler":  "terraform",
		"job_type": "terraform-deploy",
		"message":  handlerErr.Error(),
	}, resultReq.ErrorMetadata)
}

func TestExecJobStepWritesFallbackResultOnPanic(t *testing.T) {
	events := make([]string, 0, 3)
	var resultReq *models.ServiceCreateRunnerJobExecutionResultRequest
	client := &jobLoopTestClient{
		createResult: func(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
			events = append(events, "result")
			resultReq = req
			return &models.AppRunnerJobExecutionResult{ID: "result-id"}, nil
		},
		updateExecution: func(ctx context.Context, jobID, executionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error) {
			events = append(events, "status:"+string(req.Status))
			return &models.AppRunnerJobExecution{ID: executionID}, nil
		},
	}
	j := &jobLoop{apiClient: client, mw: &jobLoopTestMetrics{}}
	step := &executeJobStep{
		name:        "execute",
		handler:     &jobLoopTestHandler{name: "helm"},
		startStatus: models.AppRunnerJobExecutionStatusInDashProgress,
		fn: func(context.Context, jobs.JobHandler, *models.AppRunnerJob, *models.AppRunnerJobExecution) error {
			panic("boom")
		},
	}
	job := &models.AppRunnerJob{ID: "job-id", Type: models.AppRunnerJobTypeHelmDashChartDashDeploy}
	execution := &models.AppRunnerJobExecution{ID: "execution-id"}

	require.Panics(t, func() {
		j.execJobStep(context.Background(), zap.NewNop(), sdklog.NewLoggerProvider(), step, job, execution)
	})
	require.Equal(t, []string{
		"status:in-progress",
		"result",
		"status:failed",
	}, events)
	require.False(t, resultReq.Success)
	require.Equal(t, "execute", resultReq.ErrorMetadata["step"])
	require.Equal(t, "helm", resultReq.ErrorMetadata["handler"])
	require.Equal(t, "helm-chart-deploy", resultReq.ErrorMetadata["job_type"])
	require.Contains(t, resultReq.ErrorMetadata["message"], "panic in execute: boom")
}

func TestWriteFallbackJobExecutionResultDetachesCanceledContextAndRetries(t *testing.T) {
	attempts := 0
	client := &jobLoopTestClient{
		createResult: func(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
			attempts++
			require.NoError(t, ctx.Err())
			require.Equal(t, "job-id", jobID)
			require.Equal(t, "execution-id", executionID)
			if attempts == 1 {
				return nil, errors.New("temporary API failure")
			}
			return &models.AppRunnerJobExecutionResult{ID: "result-id"}, nil
		},
	}
	j := &jobLoop{apiClient: client}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := j.writeFallbackJobExecutionResult(
		ctx,
		&models.AppRunnerJob{ID: "job-id", Type: models.AppRunnerJobTypeTerraformDashDeploy},
		&models.AppRunnerJobExecution{ID: "execution-id"},
		"terraform",
		"execute",
		context.DeadlineExceeded,
	)
	require.NoError(t, err)
	require.Equal(t, 2, attempts)
}

func TestTerminalJobExecutionStatusDetachesCanceledContext(t *testing.T) {
	for _, withCoalescer := range []bool{false, true} {
		t.Run(fmt.Sprintf("coalescer=%t", withCoalescer), func(t *testing.T) {
			client := &jobLoopTestClient{
				updateExecution: func(ctx context.Context, jobID, executionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error) {
					require.NoError(t, ctx.Err())
					require.Equal(t, models.AppRunnerJobExecutionStatusTimedDashOut, req.Status)
					return &models.AppRunnerJobExecution{ID: executionID}, nil
				},
			}
			j := &jobLoop{apiClient: client}
			if withCoalescer {
				coalescer := newStatusCoalescer("job-id", "execution-id", zap.NewNop(), j.writeJobExecutionStatus)
				j.attachCoalescer("execution-id", coalescer)
				defer j.detachCoalescer("execution-id")
				defer coalescer.Close()
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			err := j.updateJobExecutionStatusWithDescription(
				ctx,
				"job-id",
				"execution-id",
				models.AppRunnerJobExecutionStatusTimedDashOut,
				"execution timed out",
			)
			require.NoError(t, err)
		})
	}
}
