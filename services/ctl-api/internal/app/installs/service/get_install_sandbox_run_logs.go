package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

type SandboxRunLog interface{}

// @ID GetInstallSandboxRunLogs
// @Summary	get install sandbox run logs
// @Description.markdown	get_install_sandbox_run_logs.md
// @Param			install_id	path	string	true	"install ID"
// @Param			run_id		path	string	true	"run ID"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	[]SandboxRunLog
// @Router			/v1/installs/{install_id}/sandbox-run/{run_id}/logs [get]
func (s *service) GetInstallSandboxRunLogs(ctx *gin.Context) {
	runID := ctx.Param("run_id")

	org, err := middlewares.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	logs, err := s.getSandboxRunLogs(ctx, org.ID, runID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get logs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func (s *service) getSandboxRunLogs(ctx context.Context, orgID, runID string) ([]SandboxRunLog, error) {
	logs := make([]SandboxRunLog, 0)
	ctx, cancelFn := context.WithTimeout(ctx, defaultLogPollTimeout)
	defer cancelFn()

	logClient, err := s.wpClient.GetJobStream(ctx, orgID, &gen.GetJobStreamRequest{
		JobId: runID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get job stream: %w", err)
	}

	done := false
	for {
		select {
		case <-ctx.Done():
			done = true
		default:
		}
		if done {
			break
		}

		resp, err := logClient.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if errors.Is(err, context.DeadlineExceeded) {
			break
		}
		// TODO(jm): figure out how to parse the context exceeded part from waypoint
		if err != nil {
			break
			// return nil, fmt.Errorf("unable to receive logs: %w", err)
		}

		logs = append(logs, resp.Event)
	}

	return logs, nil
}
