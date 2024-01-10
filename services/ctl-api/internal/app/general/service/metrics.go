package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/metrics"
)

type PublishMetricInput struct {
	Incr   *metrics.Incr   `json:"incr"`
	Decr   *metrics.Decr   `json:"decr"`
	Timing *metrics.Timing `json:"timing"`
	// TODO: remove this after test
	// Just making a non-functional change to create a promotion PR.
	// Generating the python SDK locally, with no changes, actually fixed the synax error.
	// This may be an edge case that only happens on the first generation, when it creates all the files from scratch.
	// If this doesn't fix it, I'll try moving this statsd stuff to the admin api.
	Event *metrics.Event `json:"event"`
}

func (m PublishMetricInput) write(mw metrics.Writer) {
	if m.Incr != nil {
		m.Incr.Write(mw)
	}
	if m.Decr != nil {
		m.Decr.Write(mw)
	}
	if m.Timing != nil {
		m.Timing.Write(mw)
	}
	if m.Event != nil {
		m.Event.Write(mw)
	}
}

// @ID PublishMetrics
// @Summary	Publish a metric from different Nuon clients for telemetry purposes.
// @Description.markdown	publish_metrics.md
// @Tags			general
// @Param			req	body	[]PublishMetricInput	true	"Input"
// @Accept			json
// @Produce		json
// @Security APIKey
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{string}	ok
// @Router			/v1/general/metrics [post]
func (s *service) PublishMetrics(ctx *gin.Context) {
	var req []PublishMetricInput
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	for _, metric := range req {
		metric.write(s.mw)
	}
	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}
