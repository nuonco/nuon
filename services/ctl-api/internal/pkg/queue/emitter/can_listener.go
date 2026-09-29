package emitter

import (
	"math/rand"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

var (
	emitterCANHistoryMax    = 1500
	emitterCANCheckInterval = 5 * time.Minute
	emitterCANStartJitter   = 60
)

const CheckCANUpdateName string = "check-can"

type CheckCANRequest struct{}

type CheckCANResponse struct {
	WorkflowType  string `json:"workflow_type"`
	Namespace     string `json:"namespace"`
	HistoryLength int    `json:"history_length"`
	HistoryMax    int    `json:"history_max"`
	Restarting    bool   `json:"restarting"`
}

func (e *emitterWorkflow) checkCANHandler(ctx workflow.Context, req *CheckCANRequest) (*CheckCANResponse, error) {
	l, _ := log.WorkflowLogger(ctx)
	restarting, resp := e.runCANCheck(ctx, l)
	if restarting {
		e.restarted = true
	}
	return resp, nil
}

func (e *emitterWorkflow) runCANCheck(ctx workflow.Context, l *zap.Logger) (bool, *CheckCANResponse) {
	info := workflow.GetInfo(ctx)
	historyLen := info.GetCurrentHistoryLength()

	if e.mw != nil {
		e.mw.Gauge(ctx, "workflow.workflow_size", float64(historyLen),
			metrics.ToTags(map[string]string{
				"namespace":     info.Namespace,
				"workflow_type": info.WorkflowType.Name,
				"is_can":        strconv.FormatBool(info.ContinuedExecutionRunID != ""),
			})...)
	}

	resp := &CheckCANResponse{
		WorkflowType:  info.WorkflowType.Name,
		Namespace:     info.Namespace,
		HistoryLength: historyLen,
		HistoryMax:    emitterCANHistoryMax,
	}

	if historyLen > emitterCANHistoryMax {
		if l != nil {
			l.Info("emitter history length exceeded threshold, triggering continue-as-new",
				zap.Int("history_length", historyLen),
				zap.Int("history_max", emitterCANHistoryMax),
			)
		}
		resp.Restarting = true
		return true, resp
	}

	if _, err := e.ensureEmitterActive(ctx); err != nil {
		if l != nil {
			l.Warn("CAN check: error checking emitter", zap.Error(err))
		}
		resp.Restarting = true
		return true, resp
	}
	if e.stopped {
		return false, resp
	}

	if err := e.ensureQueueActive(ctx); err != nil {
		if l != nil {
			l.Warn("CAN check: error checking queue", zap.Error(err))
		}
		resp.Restarting = true
		return true, resp
	}
	if e.stopped {
		return false, resp
	}

	return false, resp
}

func (e *emitterWorkflow) startCANListener(ctx workflow.Context) {
	workflow.Go(ctx, func(gCtx workflow.Context) {
		l, _ := log.WorkflowLogger(gCtx)

		jitter := time.Duration(rand.Intn(emitterCANStartJitter)) * time.Second
		if err := workflow.Sleep(gCtx, jitter); err != nil {
			return
		}

		for {
			jitterMax := int(emitterCANCheckInterval.Seconds() / 2)
			if jitterMax < 1 {
				jitterMax = 1
			}
			interval := emitterCANCheckInterval + time.Duration(rand.Intn(jitterMax))*time.Second
			if err := workflow.Sleep(gCtx, interval); err != nil {
				return
			}

			restarting, _ := e.runCANCheck(gCtx, l)
			if restarting {
				e.restarted = true
				return
			}
			if e.stopped {
				return
			}
		}
	})
}
