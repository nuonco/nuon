package stack

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/sdks/stack/client/operations"
	"github.com/nuonco/nuon/sdks/stack/models"
)

// PhoneHome reports install stack outputs, marking the operation complete. Takes
// phoneHomeURL from Config rather than composing it, and reports to that host so
// the run lands where ctl-api directed it.
func PhoneHome(ctx context.Context, opts Options, phoneHomeURL string, payload map[string]any) error {
	if strings.TrimSpace(phoneHomeURL) == "" {
		return fmt.Errorf("phone home: phone_home_url is required (read it from the stack config)")
	}
	if err := opts.validate(); err != nil {
		return fmt.Errorf("phone home: %w", err)
	}

	c, err := newClient(ctx, opts)
	if err != nil {
		return fmt.Errorf("phone home: %w", err)
	}

	return c.PhoneHome(ctx, phoneHomeURL, payload)
}

// phoneHomePath is the spec-derived path for an install's report route.
func phoneHomePath(installID string) string {
	return fmt.Sprintf("/v1/stacks/%s/phone-home", installID)
}

// PhoneHome posts to phoneHomeURL after checking it addresses this install's
// report route. The host is not checked: it varies by environment. The path is,
// because a stale capability URL carrying a phone_home_id shows up there and
// would otherwise be reported against silently.
func (c *client) PhoneHome(ctx context.Context, phoneHomeURL string, payload map[string]any) error {
	u, err := url.Parse(phoneHomeURL)
	if err != nil {
		return fmt.Errorf("phone home: parse phone_home_url %q: %w", phoneHomeURL, err)
	}

	want := phoneHomePath(c.installID)
	if got := strings.TrimSuffix(u.Path, "/"); got != want {
		return fmt.Errorf(
			"phone home: phone_home_url path is %q, expected %q — re-read phone_home_url from the stack config",
			got, want,
		)
	}

	ops, err := newOps(u.Scheme+"://"+u.Host, c.opts.HTTPClient)
	if err != nil {
		return fmt.Errorf("phone home: %w", err)
	}

	params := operations.NewPostStackPhoneHomeParamsWithContext(ctx).
		WithInstallID(c.installID).
		WithReq(models.ServiceStackPhoneHomeRequest(payload))

	if err := retry(ctx, func() error {
		_, err := ops.PostStackPhoneHome(params, c.authInfo)
		return err
	}); err != nil {
		return fmt.Errorf("phone home: %w", err)
	}

	return nil
}
