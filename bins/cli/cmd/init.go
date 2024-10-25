package cmd

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/nuonco/nuon-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	segment "github.com/segmentio/analytics-go/v3"

	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/pkg/analytics"
	"github.com/powertoolsdev/mono/pkg/errs"
)

// Construct an API client for the services to use.
func (c *cli) initAPIClient() error {
	api, err := nuon.New(
		nuon.WithValidator(c.v),
		nuon.WithAuthToken(c.cfg.APIToken),
		nuon.WithOrgID(c.cfg.OrgID),
		nuon.WithURL(c.cfg.APIURL),
	)
	if err != nil {
		return fmt.Errorf("unable to init API client: %w", err)
	}

	c.apiClient = api
	return nil
}

func (c *cli) initConfig() error {
	cfg, err := config.NewConfig(ConfigFile)
	if err != nil {
		return fmt.Errorf("unable to initialize config: %w", err)
	}

	c.cfg = cfg
	return nil
}

func (c *cli) initSentry() error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: errs.SentryMainDSN,
		// TODO(sdboyer): come up with a way of inferring from existing context that this is a dev build
		Environment: c.cfg.Env,
		Tags: map[string]string{
			"org_id":   c.cfg.OrgID,
			"platform": "cli",
		},
	})
	if err != nil {
		wrappedErr := errors.Wrap(err, "unable to initialize sentry")
		errs.ReportToSentry(wrappedErr, nil)
		return wrappedErr
	}

	return nil
}

func (c *cli) identifyFn(ctx context.Context) (*segment.Identify, error) {
	user, err := c.apiClient.GetCurrentUser(ctx)

	if err != nil {
		wrappedErr := errors.Wrap(err, "unable to get current user")
		errs.ReportToSentry(wrappedErr, nil)
		return nil, wrappedErr
	}

	return &segment.Identify{
		UserId: user.ID,
		Traits: segment.NewTraits().SetEmail(user.Email),
	}, nil
}

func (c *cli) analyticsIDFn(ctx context.Context) (string, error) {
	user, err := c.apiClient.GetCurrentUser(ctx)
	if err != nil {
		return "", errors.Wrap(err, "unable to get current user")
	}

	return user.ID, nil
}

func (c *cli) initAnalytics() error {
	// Disable zap logging when for analytics
	disabledLogger := zap.NewNop()

	ac, err := analytics.New(c.v,
		analytics.WithDisable(c.cfg.DisableTelemetry),
		analytics.WithSegmentKey(c.cfg.SegmentWriteKey),
		analytics.WithUserIDFn(c.analyticsIDFn),
		analytics.WithIdentifyFn(c.identifyFn),
		analytics.WithGroupFn(analytics.NoopGroupFn),
		analytics.WithLogger(disabledLogger),
		analytics.WithProperties(map[string]interface{}{
			"platform": "cli",
			"env":      c.cfg.Env,
		}),
	)
	if err != nil {
		return errors.Wrap(err, "unable to get analytics writer")
	}

	c.analyticsClient = ac
	return nil
}
