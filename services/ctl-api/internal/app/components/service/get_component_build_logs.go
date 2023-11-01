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
	defaultLogPollTimeout time.Duration = time.Second * 20
)

type BuildLog interface{}

//	@BasePath	/v1/installs
//
// Get install build logs
//
//	@Summary	get install build logs
//	@Schemes
//	@Description	get install build logs
//	@Param			component_id	path	string	true	"component ID"
//	@Param			build_id		path	string	true	"build ID"
//	@Tags			components
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object}	[]BuildLog
//	@Router			/v1/components/{component_id}/builds/{build_id}/logs [get]
func (s *service) GetComponentBuildLogs(ctx *gin.Context) {
	buildID := ctx.Param("build_id")

	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	logs, err := s.getLogs(ctx, org.ID, buildID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get logs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func (s *service) getLogs(ctx context.Context, orgID, buildID string) ([]BuildLog, error) {
	logs := make([]BuildLog, 0)
	ctx, cancelFn := context.WithTimeout(ctx, defaultLogPollTimeout)
	defer cancelFn()

	logClient, err := s.wpClient.GetJobStream(ctx, orgID, &gen.GetJobStreamRequest{
		JobId: fmt.Sprintf("build-%s", buildID),
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
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("unable to receive logs: %w", err)
		}

		logs = append(logs, resp.Event)
	}

	return logs, nil
}
