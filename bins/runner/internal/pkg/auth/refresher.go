package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

const (
	refreshFetchTimeout = 30 * time.Second
	minRefreshDelay     = 10 * time.Second
	minRefreshRetry     = 30 * time.Second
	maxRefreshRetry     = 5 * time.Minute
)

type RefresherParams struct {
	fx.In

	Cfg       *runnerconfig.Config
	Token     *Token
	APIClient nuonrunner.Client
	L         *zap.Logger `name:"system"`
	LC        fx.Lifecycle
}

// Refresher re-authenticates with the cloud identity at half of the remaining
// token lifetime, so a short-lived runner token never expires under a running
// process.
type Refresher struct {
	cfg       *runnerconfig.Config
	apiClient nuonrunner.Client
	l         *zap.Logger
	expiresAt time.Time

	fetch func(context.Context, *runnerconfig.Config) (*Token, error)

	ctx      context.Context
	cancelFn func()
	wg       *conc.WaitGroup
}

func NewRefresher(params RefresherParams) *Refresher {
	ctx, cancelFn := context.WithCancel(context.Background())

	r := &Refresher{
		cfg:       params.Cfg,
		apiClient: params.APIClient,
		l:         params.L,
		expiresAt: params.Token.ExpiresAt,
		fetch:     fetchRefreshedToken,
		ctx:       ctx,
		cancelFn:  cancelFn,
		wg:        conc.NewWaitGroup(),
	}

	if r.expiresAt.IsZero() {
		return r
	}

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			r.wg.Go(func() { r.loop(r.ctx) })
			return nil
		},
		OnStop: func(context.Context) error {
			r.cancelFn()
			r.wg.Wait()
			return nil
		},
	})

	return r
}

func fetchRefreshedToken(ctx context.Context, cfg *runnerconfig.Config) (*Token, error) {
	result, err := fetch(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Token{Value: result.Token, ExpiresAt: result.ExpiresAt}, nil
}

func (r *Refresher) loop(ctx context.Context) {
	retry := minRefreshRetry
	wait := refreshDelay(time.Now(), r.expiresAt)

	for {
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}

		if err := r.refresh(ctx); err != nil {
			r.l.Warn("unable to refresh runner token",
				zap.Error(err),
				zap.Time("expires_at", r.expiresAt),
				zap.Duration("retry_in", retry))
			wait = retry
			retry = min(retry*2, maxRefreshRetry)
			continue
		}

		if r.expiresAt.IsZero() {
			r.l.Info("refreshed runner token has no expiry, stopping refresh")
			return
		}

		retry = minRefreshRetry
		wait = refreshDelay(time.Now(), r.expiresAt)
	}
}

func (r *Refresher) refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, refreshFetchTimeout)
	defer cancel()

	token, err := r.fetch(ctx, r.cfg)
	if err != nil {
		return fmt.Errorf("unable to fetch runner token: %w", err)
	}

	r.apiClient.SetAuthToken(token.Value)
	r.cfg.SetAPIToken(token.Value)
	r.expiresAt = token.ExpiresAt

	r.l.Info("refreshed runner token", zap.Time("expires_at", token.ExpiresAt))
	return nil
}

func refreshDelay(now, expiresAt time.Time) time.Duration {
	return max(expiresAt.Sub(now)/2, minRefreshDelay)
}
