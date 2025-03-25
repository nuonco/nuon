package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
)

type CreateRunnerJobExecutionOutputsRequest struct {
	Outputs map[string]interface{} `json:"outputs"`
}

//	@ID						CreateRunnerJobExecutionOutputs
//	@Summary				create a runner job execution outputs
//	@Description.markdown	create_runner_job_execution_outputs.md
//	@Param					req						body	CreateRunnerJobExecutionOutputsRequest	true	"Input"
//	@Param					runner_job_id			path	string									true	"runner job ID"
//	@Param					runner_job_execution_id	path	string									true	"runner job execution ID"
//	@Tags					runners/runner
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				201	{object}	app.RunnerJobExecutionOutputs
//	@Router					/v1/runner-jobs/{runner_job_id}/executions/{runner_job_execution_id}/outputs [POST]
func (s *service) CreateRunnerJobExecutionOutputs(ctx *gin.Context) {
	runnerJobExecutionID := ctx.Param("runner_job_execution_id")
	runnerJobID := ctx.Param("runner_job_id")

	jobExecution, err := s.getRunnerJobExecution(ctx, runnerJobID, runnerJobExecutionID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := authz.CanCreate(ctx, jobExecution.OrgID); err != nil {
		ctx.Error(err)
		return
	}

	var req CreateRunnerJobExecutionOutputsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(errors.Wrap(err, "unable to parse request"))
		return
	}

	outputs, err := s.createRunnerJobExecutionOutputs(ctx, runnerJobExecutionID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, outputs)
}

func (s *service) createRunnerJobExecutionOutputs(ctx context.Context, runnerJobExecutionID string, req *CreateRunnerJobExecutionOutputsRequest) (*app.RunnerJobExecutionOutputs, error) {
	byts, err := json.Marshal(req.Outputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert outputs to json")
	}

	obj := app.RunnerJobExecutionOutputs{
		RunnerJobExecutionID: runnerJobExecutionID,
		Outputs:              byts,
	}

	res := s.db.WithContext(ctx).
		Create(&obj)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to write runner job execution result")
	}

	return &obj, nil
}
