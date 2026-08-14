package audit

import (
	"context"
	"errors"
	"net"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/settings"
)

const (
	routePollInterval  = time.Second
	routeRetryInterval = 30 * time.Second
	routeDialTimeout   = 50 * time.Millisecond
)

var Module = fx.Options(fx.Provide(New, NewClient), fx.Invoke(func(*Client) {}))

// LocalRouteLifecycle orders Client shutdown before an in-process route owner without coupling audit to its implementation.
type LocalRouteLifecycle interface {
	AuditRouteLifecycle()
}

type ClientParams struct {
	fx.In

	Lifecycle      fx.Lifecycle
	Settings       *settings.Settings
	Logger         *zap.Logger `name:"system"`
	Writer         *Writer
	RouteLifecycle LocalRouteLifecycle `optional:"true"`
}

type Client struct {
	local            bool
	installID        string
	logger           *zap.Logger
	writer           clientWriter
	cancel           context.CancelFunc
	done             chan struct{}
	routeAvailableFn func() bool
	pollInterval     time.Duration
	retryInterval    time.Duration
}

type clientWriter interface {
	Enable() error
	Disable()
	ProcessStopping(context.Context, string, string) error
}

func NewClient(params ClientParams) *Client {
	c := &Client{
		local:            params.Settings.Cfg.IsNuonctl,
		installID:        params.Settings.Metadata["install.id"],
		logger:           params.Logger,
		writer:           params.Writer,
		done:             make(chan struct{}),
		routeAvailableFn: syncRouteAvailable,
		pollInterval:     routePollInterval,
		retryInterval:    routeRetryInterval,
	}
	params.Lifecycle.Append(fx.Hook{OnStart: c.start, OnStop: c.stop})
	return c
}

func (c *Client) start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go c.run(ctx)
	return nil
}

func (c *Client) stop(ctx context.Context) error {
	if c.installID != "" && c.writer != nil {
		if err := c.writer.ProcessStopping(ctx, "graceful_shutdown", "fx_lifecycle"); err != nil && !errors.Is(err, ErrUnavailable) {
			c.logger.Warn("customer audit shutdown event export failed", zap.Error(err))
		}
	}
	if c.cancel != nil {
		c.cancel()
		select {
		case <-c.done:
		case <-ctx.Done():
		}
	}
	return nil
}

func (c *Client) run(ctx context.Context) {
	defer close(c.done)
	if c.local || c.installID == "" || c.writer == nil {
		return
	}
	defer c.writer.Disable()

	enabled := false
	var retryAt time.Time
	for {
		available := c.routeAvailableFn()
		switch {
		case !available:
			if enabled {
				c.writer.Disable()
				enabled = false
				c.logger.Info("runner customer audit client disabled", zap.String("audit_export.reason", "synchronous route unavailable"))
			}
			retryAt = time.Time{}
		case !enabled && (retryAt.IsZero() || !time.Now().Before(retryAt)):
			if err := c.writer.Enable(); err != nil {
				c.writer.Disable()
				retryAt = time.Now().Add(c.retryInterval)
				c.logger.Warn("customer audit startup event was not acknowledged by the synchronous route",
					zap.Bool("audit_export.delivery_verified", false),
					zap.String("audit_export.failure_type", failureType(err)),
				)
			} else {
				enabled = true
				retryAt = time.Time{}
				c.logger.Info("runner customer audit client enabled")
			}
		}
		if !wait(ctx, c.pollInterval) {
			return
		}
	}
}

func syncRouteAvailable() bool {
	conn, err := net.DialTimeout("tcp", SyncRouteAddress, routeDialTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func wait(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func failureType(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, ErrUnavailable):
		return "unavailable"
	default:
		return "failure"
	}
}
