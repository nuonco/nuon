package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						AdminGetLogStream
//	@Summary				get a log stream
//	@Description.markdown	admin_get_log_stream.md
//	@Param					log_stream_id	path	string	true	"log stream or owner ID"
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{object}	app.LogStream
//	@Router					/v1/log-streams/{log_stream_id} [GET]
func (s *service) AdminGetLogStream(ctx *gin.Context) {
	logStreamID := ctx.Param("log_stream_id")

	ls, err := s.adminGetLogStream(ctx, logStreamID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get log stream: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, ls)
}

func (s *service) adminGetLogStream(ctx *gin.Context, logStreamID string) (*app.LogStream, error) {
	logStream := app.LogStream{}
	res := s.db.WithContext(ctx).
		Where("owner_id = ?", logStreamID).
		Or("id = ?", logStreamID).
		First(&logStream)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get log stream: %w", res.Error)
	}

	return &logStream, nil
}
