package revokeunusedtokens

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "revoke_unused_tokens"

type Signal struct {
	RunnerID string `json:"runner_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}

	if _, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID); err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	resp, err := activities.AwaitRevokeUnusedTokens(ctx, activities.RevokeUnusedTokensRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to revoke unused tokens")
	}

	l, err := log.WorkflowLogger(ctx)
	if err == nil {
		l.Info("revoked unused runner tokens",
			zap.String("runner_id", s.RunnerID),
			zap.Int64("revoked_count", resp.RevokedCount))
	}

	return nil
}
