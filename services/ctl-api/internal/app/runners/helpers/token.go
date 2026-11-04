package helpers

import (
	"context"
	"slices"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	defaultRunnerTokenTimeout   time.Duration = time.Hour * 24 * 90
	bootstrapRunnerTokenTimeout time.Duration = time.Hour * 2

	minRevokableTokenAge time.Duration = time.Hour * 12
)

var ErrTokenCleanupNotReady = errors.New("token cleanup not ready")

var tokenHoldingRunnerProcessStatuses = append(
	app.ActiveRunnerProcessStatuses(),
	app.RunnerProcessStatusOffline,
)

func (a *Helpers) CreateToken(ctx context.Context, runnerID string) (*app.Token, error) {
	email := account.ServiceAccountEmail(runnerID)

	token, err := a.acctClient.CreateToken(ctx, email, defaultRunnerTokenTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	return token, nil
}

func (a *Helpers) CreateBootstrapToken(ctx context.Context, runnerID string) (*app.Token, error) {
	email := account.ServiceAccountEmail(runnerID)

	token, err := a.acctClient.CreateToken(ctx, email, bootstrapRunnerTokenTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create bootstrap token")
	}

	return token, nil
}

func (a *Helpers) InvalidateOldTokens(ctx context.Context, runnerID string) (int64, error) {
	email := account.ServiceAccountEmail(runnerID)

	count, err := a.acctClient.InvalidateOldTokens(ctx, email)
	if err != nil {
		return 0, errors.Wrap(err, "unable to invalidate old tokens")
	}

	return count, nil
}

func (a *Helpers) RevokeReplacedToken(ctx context.Context, runnerID, tokenID string) (int64, error) {
	processes, err := a.runnerProcesses(ctx, runnerID, tokenProtectedRunnerProcessStatuses())
	if err != nil {
		return 0, err
	}
	keep, err := tokensToKeep(processes)
	if err != nil {
		if errors.Is(err, ErrTokenCleanupNotReady) {
			return 0, nil
		}
		return 0, err
	}
	if slices.Contains(keep, tokenID) {
		return 0, nil
	}

	email := account.ServiceAccountEmail(runnerID)
	count, err := a.acctClient.RevokeTokenCreatedBefore(ctx, email, tokenID, time.Now().Add(-minRevokableTokenAge))
	if err != nil {
		return 0, errors.Wrap(err, "unable to revoke replaced token")
	}

	return count, nil
}

func (a *Helpers) RevokeUnusedTokens(ctx context.Context, runnerID string) (int64, error) {
	processes, err := a.runnerProcesses(ctx, runnerID, tokenProtectedRunnerProcessStatuses())
	if err != nil {
		return 0, err
	}

	keep, err := tokensToKeep(processes)
	if err != nil {
		return 0, err
	}

	email := account.ServiceAccountEmail(runnerID)
	count, err := a.acctClient.RevokeTokensExcept(ctx, email, keep, time.Now().Add(-minRevokableTokenAge))
	if err != nil {
		return 0, errors.Wrap(err, "unable to revoke unused tokens")
	}

	return count, nil
}

func tokensToKeep(processes []app.RunnerProcess) ([]string, error) {
	pendingStatus := app.RunnerProcessStatus(app.StatusPending)
	live := make([]app.RunnerProcess, 0, len(processes))
	pendingTokens := make([]string, 0)
	for _, process := range processes {
		if process.ProcessStatus() == pendingStatus {
			if process.TokenID != "" {
				pendingTokens = append(pendingTokens, process.TokenID)
			}
			continue
		}
		live = append(live, process)
	}
	if len(live) == 0 {
		return nil, errors.Wrap(ErrTokenCleanupNotReady, "runner has no live processes")
	}

	keep := make([]string, 0, len(live)+len(pendingTokens))
	for _, process := range live {
		if process.TokenID == "" {
			return nil, errors.Wrapf(ErrTokenCleanupNotReady, "runner process %s has not recorded its token yet", process.ID)
		}
		keep = append(keep, process.TokenID)
	}

	return append(keep, pendingTokens...), nil
}

func TokenHoldingRunnerProcessStatuses() []app.RunnerProcessStatus {
	return append([]app.RunnerProcessStatus{}, tokenHoldingRunnerProcessStatuses...)
}

func tokenProtectedRunnerProcessStatuses() []app.RunnerProcessStatus {
	return append(
		append([]app.RunnerProcessStatus{}, tokenHoldingRunnerProcessStatuses...),
		app.RunnerProcessStatus(app.StatusPending),
	)
}

func (a *Helpers) runnerProcesses(ctx context.Context, runnerID string, statuses []app.RunnerProcessStatus) ([]app.RunnerProcess, error) {
	var processes []app.RunnerProcess
	if err := a.db.WithContext(ctx).
		Where(app.RunnerProcess{RunnerID: runnerID}).
		Where("composite_status->>'status' IN ?", statuses).
		Find(&processes).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get runner processes")
	}

	return processes, nil
}
