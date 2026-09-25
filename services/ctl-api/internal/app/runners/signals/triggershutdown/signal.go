package triggershutdown

import (
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "trigger_shutdown"

const processQueuePrefix = "runner-process-"

type Signal struct {
	RunnerID    string `json:"runner_id"`
	ProcessType string `json:"process_type"`
	// ProcessID pins the shutdown to the process whose uptime emitter fired.
	// Templates created before it existed leave it empty.
	ProcessID string `json:"process_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.ProcessType == "" {
		return errors.New("process_type is required")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	processID := s.ProcessID
	if processID == "" {
		var err error
		processID, err = s.processIDFromQueue(ctx)
		if err != nil {
			return err
		}
	}
	if processID == "" {
		// A stale emitter must never shut down whichever process happens to be current.
		l.Warn("trigger_shutdown without a process id, skipping", "runner_id", s.RunnerID)
		return nil
	}

	process, err := activities.AwaitGetRunnerProcessByProcessID(ctx, processID)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "unable to get process")
	}

	if process.RunnerID != s.RunnerID || process.ProcessStatus() != app.RunnerProcessStatusActive {
		return nil
	}

	_, err = activities.AwaitCreateRunnerProcessShutdown(ctx, activities.CreateRunnerProcessShutdownRequest{
		RunnerProcessID: process.ID,
		Type:            app.RunnerProcessShutdownTypeGraceful,
		CompositeStatus: app.CompositeStatus{
			Status:                 app.Status(app.RunnerProcessShutdownStatusRequested),
			StatusHumanDescription: "uptime threshold exceeded",
			CreatedAtTS:            workflow.Now(ctx).Unix(),
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to create shutdown for process")
	}

	return nil
}

// processIDFromQueue recovers the target for legacy templates: the emitter
// always lives on that process's own runner-process-<id> queue.
func (s *Signal) processIDFromQueue(ctx workflow.Context) (string, error) {
	queueID := cctx.QueueIDFromContext(ctx)
	if queueID == "" {
		return "", nil
	}
	q, err := queueclient.AwaitGetQueue(ctx, queueID)
	if err != nil {
		return "", errors.Wrap(err, "unable to get queue")
	}
	if !strings.HasPrefix(q.Name, processQueuePrefix) {
		return "", nil
	}
	return strings.TrimPrefix(q.Name, processQueuePrefix), nil
}
