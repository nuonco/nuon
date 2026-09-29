package queue

import (
	"math/rand"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

const (
	canDefaultHintPeriod        = 3 * time.Minute
	canStartJitter              = 60
	canDefaultHistoryMax        = 10000
	canDefaultTerminateOverhead = 5000
)

const CheckCANUpdateName string = "check-can"

type CheckCANRequest struct{}

type CheckCANResponse struct {
	WorkflowType  string `json:"workflow_type"`
	Namespace     string `json:"namespace"`
	HistoryLength int    `json:"history_length"`
	HistoryMax    int    `json:"history_max"`
	HintRequested bool   `json:"hint_requested"`
	Restarting    bool   `json:"restarting"`
}

func (q *queue) checkCANHandler(ctx workflow.Context, req *CheckCANRequest) (*CheckCANResponse, error) {
	l, _ := log.WorkflowLogger(ctx)
	if q.activeWorkers > 0 {
		info := workflow.GetInfo(ctx)
		historyMax := canDefaultHistoryMax
		if q.cfg != nil && q.cfg.QueueContinueAsNewHistoryMax > 0 {
			historyMax = q.cfg.QueueContinueAsNewHistoryMax
		}
		return &CheckCANResponse{
			WorkflowType:  info.WorkflowType.Name,
			Namespace:     info.Namespace,
			HistoryLength: info.GetCurrentHistoryLength(),
			HistoryMax:    historyMax,
		}, nil
	}
	restarting, resp := q.runCANCheck(ctx, l)
	if restarting {
		q.setStatus(ctx, l, QueueStatusRestartAccepted)
		q.restarted = true
	}
	return resp, nil
}

func (q *queue) runCANCheck(ctx workflow.Context, l *zap.Logger) (bool, *CheckCANResponse) {
	info := workflow.GetInfo(ctx)
	historyLen := info.GetCurrentHistoryLength()
	if q.mw != nil {
		info := workflow.GetInfo(ctx)
		q.mw.Gauge(ctx, "workflow.workflow_size", float64(historyLen),
			metrics.ToTags(map[string]string{
				"namespace":     info.Namespace,
				"workflow_type": info.WorkflowType.Name,
				"is_can":        strconv.FormatBool(info.ContinuedExecutionRunID != ""),
			})...)
	}

	historyMax := canDefaultHistoryMax
	if q.cfg != nil && q.cfg.QueueContinueAsNewHistoryMax > 0 {
		historyMax = q.cfg.QueueContinueAsNewHistoryMax
	}

	resp := &CheckCANResponse{
		WorkflowType:  info.WorkflowType.Name,
		Namespace:     info.Namespace,
		HistoryLength: historyLen,
		HistoryMax:    historyMax,
	}

	if historyLen > historyMax {
		if l != nil {
			l.Info("history length exceeded threshold, triggering continue-as-new",
				zap.Int("history_length", historyLen))
		}
		resp.Restarting = true
		return true, resp
	}

	requested, err := activities.LocalAwaitCheckCANRequested(ctx, activities.CheckCANRequestedRequest{
		QueueID: q.queueID,
	})
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			if l != nil {
				l.Warn("queue not found during CAN check, stopping workflow", zap.String("queue-id", q.queueID))
			}
			q.stopped = true
			return false, resp
		}

		if l != nil {
			l.Warn("CAN check failed, restarting workflow to recover", zap.Error(err))
		}
		resp.Restarting = true
		return true, resp
	}

	resp.HintRequested = requested
	if requested {
		if l != nil {
			l.Info("continue-as-new requested via metadata, clearing hint")
		}
		if clearErr := activities.LocalAwaitClearCANRequested(ctx, activities.ClearCANRequestedRequest{
			QueueID: q.queueID,
		}); clearErr != nil && l != nil {
			l.Warn("unable to clear restart_hint", zap.Error(clearErr))
		}
		resp.Restarting = true
		return true, resp
	}

	return false, resp
}

func (q *queue) startCANListener(ctx workflow.Context) {
	workflow.Go(ctx, func(gCtx workflow.Context) {
		l, _ := log.WorkflowLogger(gCtx)

		jitter := time.Duration(rand.Intn(canStartJitter)) * time.Second
		if err := workflow.Sleep(gCtx, jitter); err != nil {
			return
		}

		for {
			hintPeriod := canDefaultHintPeriod
			if q.cfg != nil && q.cfg.QueueContinueAsNewHintPeriod > 0 {
				hintPeriod = q.cfg.QueueContinueAsNewHintPeriod
			}
			jitterMax := int(hintPeriod.Seconds() / 2)
			if jitterMax < 1 {
				jitterMax = 1
			}
			interval := hintPeriod + time.Duration(rand.Intn(jitterMax))*time.Second
			if err := workflow.Sleep(gCtx, interval); err != nil {
				return
			}

			restarting, _ := q.runCANCheck(gCtx, l)
			if restarting {
				q.setStatus(gCtx, l, QueueStatusRestartAccepted)
				q.restarted = true
				return
			}
			if q.stopped {
				return
			}
		}
	})
}
