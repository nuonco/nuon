package activities

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRunnerJobRequest struct {
	RunnerJobOwnerID string `validate:"required"`
}

type GetRunnerJobExecutionRequest struct {
	RunnerJobID string `validate:"required"`
}

type GetRunnerJobExecutionResultRequest struct {
	RunnerJobExecutionID string `validate:"required"`
}

func (a *Activities) getRunnerJobExecution(ctx context.Context, req GetRunnerJobExecutionRequest) (*app.RunnerJobExecution, error) {
	var runnerJobExecution app.RunnerJobExecution

	res := a.db.WithContext(ctx).Where(app.RunnerJobExecution{
		RunnerJobID: req.RunnerJobID,
	}).
		Order("created_at DESC").
		First(&runnerJobExecution)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job execution, runner job id: %s: %w", req.RunnerJobID, res.Error)
	}

	return &runnerJobExecution, nil
}

func (a *Activities) getRunnerJobExecutionResult(ctx context.Context, req GetRunnerJobExecutionResultRequest) (*app.RunnerJobExecutionResult, error) {
	var runnerJobExecutionResult app.RunnerJobExecutionResult

	res := a.db.WithContext(ctx).Where(app.RunnerJobExecutionResult{
		RunnerJobExecutionID: req.RunnerJobExecutionID,
	}).
		Order("created_at DESC").
		First(&runnerJobExecutionResult)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job execution result, runner job execution id: %s: %w", req.RunnerJobExecutionID, res.Error)
	}

	return &runnerJobExecutionResult, nil
}

func (a *Activities) getRunnerJob(ctx context.Context, req *GetRunnerJobRequest) (*app.RunnerJob, error) {
	var runnerJob app.RunnerJob

	res := a.db.WithContext(ctx).Where(app.RunnerJob{
		OwnerID: req.RunnerJobOwnerID,
	}).
		Order("created_at DESC").
		First(&runnerJob)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job jobOwnerID: %s: %w", req.RunnerJobOwnerID, res.Error)
	}

	return &runnerJob, nil
}

type GetApprovalPlanRequest struct {
	StepTargetID string `validate:"required"`
}

// @temporal-gen activity
// @max-retries 1
func (a *Activities) GetApprovalPlan(ctx context.Context, req GetApprovalPlanRequest) (*ApprovalPlan, error) {
	runnerJob, err := a.getRunnerJob(ctx, &GetRunnerJobRequest{RunnerJobOwnerID: req.StepTargetID})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to fetch runner job, step target id: %s", req.StepTargetID))
	}

	runnerJobExecution, err := a.getRunnerJobExecution(ctx, GetRunnerJobExecutionRequest{
		RunnerJobID: runnerJob.ID,
	})
	if err != nil {
		return nil, err
	}

	runnerJobExecutionResult, err := a.getRunnerJobExecutionResult(ctx, GetRunnerJobExecutionResultRequest{
		RunnerJobExecutionID: runnerJobExecution.ID,
	})
	if err != nil {
		return nil, err
	}

	// we're only using content display currently since we're only dealing with terraform and sandbox plans
	decompressedContentDisplay, err := a.decompressRunnerJobExecutionResult(runnerJobExecutionResult.ContentsDisplayGzip)
	if err != nil {
		return nil, err
	}

	plan := ApprovalPlan{
		RunnerJobType: runnerJob.Type,
		PlanContents:  decompressedContentDisplay,
	}

	return &plan, nil
}

func (a *Activities) decompressRunnerJobExecutionResult(b []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, errors.Wrap(err, "unable to decompress plan contents, failed to read contents")
	}
	defer gz.Close()

	var output bytes.Buffer
	n, err := io.Copy(&output, gz)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decompress plan contentsn ")
	}

	if n == 0 {
		return nil, errors.Wrap(err, "decompressed file size 0")
	}

	ob := output.Bytes()

	return ob, nil
}
