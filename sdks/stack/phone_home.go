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
// report route. Only the route suffix is checked: the host varies by environment,
// and the runner api is served under a path prefix in some of them. The suffix is
// checked because a stale capability URL carrying a phone_home_id shows up there
// and would otherwise be reported against silently.
func (c *client) PhoneHome(ctx context.Context, phoneHomeURL string, payload map[string]any) error {
	u, err := url.Parse(phoneHomeURL)
	if err != nil {
		return fmt.Errorf("phone home: parse phone_home_url %q: %w", phoneHomeURL, err)
	}

	want := phoneHomePath(c.installID)
	got := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(got, want) {
		return fmt.Errorf(
			"phone home: phone_home_url path is %q, expected it to end with %q — re-read phone_home_url from the stack config",
			got, want,
		)
	}

	// Whatever precedes the route is the api's base path, and has to be kept.
	base := u.Scheme + "://" + u.Host + strings.TrimSuffix(got, want)

	ops, err := newOps(base, c.opts.HTTPClient)
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
