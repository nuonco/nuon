package activities

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	ghApiTimeout = 5 * time.Second

	// DefaultAWSPhoneHomeScript is pinned to a tag rather than refs/heads/main on
	// purpose. Under a branch ref, any commit to that file in nuonco/runner would ship
	// itself to every org on their next stack regeneration — including orgs with
	// phone-home auth off, who have no reason to take the change. Bumping this is a
	// deliberate act; use PhoneHomeScriptURL to try a script out first.
	DefaultAWSPhoneHomeScript = "https://raw.githubusercontent.com/nuonco/runner/refs/tags/aws-v0.1.10/scripts/aws/phonehome.py"
)

type GetPhoneHomeScriptRequest struct {
	// URL is the app's AppRunnerConfig.PhoneHomeScriptURL. Empty falls back to the
	// environment override and then to DefaultAWSPhoneHomeScript.
	URL string `json:"url,omitempty"`
}

// FetchPhoneHomeScriptRaw is the single resolution + fetch path for the phone-home
// Lambda source. It is package-level so non-Temporal callers (e.g. app bundle
// compilation) share the exact precedence and pin that install stack versions get.
//
// Resolution is app override, then environment override, then the pinned default. The
// precedence lives here rather than in the callers so the generation paths cannot
// disagree about which script an install gets.
func FetchPhoneHomeScriptRaw(ctx context.Context, appURL, envURL string) ([]byte, error) {
	url := DefaultAWSPhoneHomeScript
	switch {
	case appURL != "":
		url = appURL
	case envURL != "":
		url = envURL
	}

	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request for phone-home script")
	}

	ctx, cancel := context.WithTimeout(ctx, ghApiTimeout)
	defer cancel()

	r = r.WithContext(ctx)
	client := http.DefaultClient

	resp, err := client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch phone-home script")
	}
	defer resp.Body.Close()

	// A typo'd override otherwise renders GitHub's 404 page into the template as the
	// Lambda's source, which fails at CreateStack in the customer's account instead of
	// here.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unable to fetch phone-home script from %s: HTTP %d", url, resp.StatusCode)
	}

	byts, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body of phone-home script")
	}

	return byts, nil
}

// @temporal-gen-v2 activity
func (a *Activities) GetPhoneHomeScriptRaw(ctx context.Context, req *GetPhoneHomeScriptRequest) ([]byte, error) {
	var appURL string
	if req != nil {
		appURL = req.URL
	}
	return FetchPhoneHomeScriptRaw(ctx, appURL, a.cfg.PhoneHomeScriptURL)
}
