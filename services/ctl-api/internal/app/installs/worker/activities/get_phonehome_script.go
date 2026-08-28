package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

const DefaultAWSPhoneHomeScript = cloudformation.DefaultAWSPhoneHomeScript

type GetPhoneHomeScriptRequest struct {
	// URL is the app's AppRunnerConfig.PhoneHomeScriptURL. Empty falls back to the
	// environment override and then to DefaultAWSPhoneHomeScript.
	URL string `json:"url,omitempty"`
}

// GetPhoneHomeScriptRaw fetches the phone-home Lambda source to embed in the rendered
// stack template.
//
// Resolution is app override, then environment override, then the pinned default. The
// precedence lives here rather than in the two callers so both generation paths cannot
// disagree about which script an install gets.
//
// @temporal-gen-v2 activity
func (a *Activities) GetPhoneHomeScriptRaw(ctx context.Context, req *GetPhoneHomeScriptRequest) ([]byte, error) {
	appURL := ""
	if req != nil {
		appURL = req.URL
	}
	return cloudformation.FetchPhoneHomeScript(ctx, appURL, a.cfg.PhoneHomeScriptURL)
}
