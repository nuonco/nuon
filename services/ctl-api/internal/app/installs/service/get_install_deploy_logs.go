package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

const (
	defaultLogPollTimeout time.Duration = time.Second * 2
)

type DeployLog interface{}

// @ID GetInstallDeployLogs
// @Summary	get install deploy logs
// @Description.markdown	get_install_deploy_logs.md
// @Param			install_id	path	string	true	"install ID"
// @Param			deploy_id	path	string	true	"deploy ID"
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
// @Success		200				{object}	[]DeployLog
// @Router			/v1/installs/{install_id}/deploys/{deploy_id}/logs [get]
func (s *service) GetInstallDeployLogs(ctx *gin.Context) {
	deployID := ctx.Param("deploy_id")

	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	logs, err := s.getLogs(ctx, org.ID, deployID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get logs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func (s *service) getLogs(ctx context.Context, orgID, deployID string) ([]DeployLog, error) {
	logs := make([]DeployLog, 0)
	ctx, cancelFn := context.WithTimeout(ctx, defaultLogPollTimeout)
	defer cancelFn()

	logClient, err := s.wpClient.GetJobStream(ctx, orgID, &gen.GetJobStreamRequest{
		JobId: fmt.Sprintf("deploy-%s", deployID),
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
			//return nil, fmt.Errorf("unable to receive logs: %w", err)
		}

		logs = append(logs, resp.Event)
	}

	return logs, nil
}
