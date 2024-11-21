package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @ID AdminGetLogStreamLogs
// @Summary	get a log stream's logs
// @Description.markdown admin_get_log_stream_logs.md
// @Param			log_stream_id	path	string	true	"log stream or owner ID"
// @Tags runners/admin
// @Security AdminEmail
// @Accept			json
// @Produce		json
// @Success		200				{object}	[]app.OtelLogRecord
// @Router			/v1/log-streams/{log_stream_id}/logs [GET]
func (s *service) AdminGetLogStreamLogs(ctx *gin.Context) {
	logStreamID := ctx.Param("log_stream_id")

	ls, err := s.adminGetLogStream(ctx, logStreamID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get log stream: %w", err))
		return
	}

	before := time.Now().UTC().UnixNano()
	logs, headers, err := s.getLogStreamLogs(ctx, ls.ID, before)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to read runner logs: %w", err))
		return
	}
	for key, value := range headers {
		ctx.Header(key, value)
	}

	ctx.JSON(http.StatusOK, logs)
}
