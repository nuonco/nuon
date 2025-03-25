package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						GetLogStream
//	@Summary				get a log stream
//	@Description.markdown	get_log_stream.md
//	@Param					log_stream_id	path	string	true	"log stream ID"
//	@Tags					runners
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				200	{object}	app.LogStream
//	@Router					/v1/log-streams/{log_stream_id} [get]
func (s *service) GetLogStream(ctx *gin.Context) {
	logStreamID := ctx.Param("log_stream_id")

	logStream, err := s.getLogStream(ctx, logStreamID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, logStream)
}

func (s *service) getLogStream(ctx context.Context, logStreamID string) (*app.LogStream, error) {
	logStream := app.LogStream{}
	res := s.db.WithContext(ctx).
		First(&logStream, "id = ?", logStreamID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get log stream")
	}

	return &logStream, nil
}
