package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/errors"
	"github.com/getsentry/sentry-go"
	"github.com/nuonco/nuon-go"
	"go.uber.org/zap"

	segment "github.com/segmentio/analytics-go/v3"

	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/bins/cli/internal/version"
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

func (c *cli) checkCLIVersion() error {
	if version.Version == "development" {
		// ignore this check if the cli is a dev build
		return nil
	}

	resp, err := stdhttp.Get(fmt.Sprintf("%s/version", c.cfg.APIURL))
	if err != nil {
		return errors.Wrap(err, "unable to get Nuon API version for cli version check")
	}
	defer resp.Body.Close()
	byt, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "unable to Nuon API version response response body")
	}
	body := make(map[string]string)
	if err := json.Unmarshal(byt, &body); err != nil {
		return errors.Wrap(err, "unable to unmarshal Nuon API version response body")
	}
	vstr, has := body["version"]
	if !has {
		return errors.New("Nuon API version response body does not contain expected version field")
	}

	if vstr == "development" || body["git_ref"] == "local" {
		// ignore this check if the api is a dev build
		return nil
	}

	apiv, err := semver.NewVersion(vstr)
	if err != nil {
		return errors.Wrap(err, "unable to parse Nuon API version to semver")
	}
	cliv, err := semver.NewVersion(version.Version)
	if err != nil {
		return errors.Wrap(err, "unable to parse cli version to semver")
	}

	switch apiv.Compare(cliv) {
	case -1:
		return errors.Newf("should be unreachable - cli version (%s) is newer than API version (%s)", cliv, apiv)
	case 1:
		bumped := cliv.IncPatch()
		if apiv.Compare(&bumped) == 0 {
			// TODO(sdboyer) make this more visible
			fmt.Print("\nA new version of the Nuon CLI is available.\n\nSee https://docs.nuon.co/cli for information on updating.\n\n")
			return nil
		}
		return errors.Newf("Your Nuon CLI (%s) is too out of date (latest is %s). See https://docs.nuon.co/cli for information on updating.\n", cliv, apiv)
	}

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
		Dsn:         c.cfg.SentryDSN,
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

func (c *cli) initUser() error {
	if c.cfg.APIToken == "" {
		return nil
	}
	user, err := c.apiClient.GetCurrentUser(c.ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get current user")
	}

	c.cfg.UserID = user.ID
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
