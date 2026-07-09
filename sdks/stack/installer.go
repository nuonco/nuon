package stack

import (
	"context"
	"fmt"
)

// FetchConfig reads the rendered install-stack configuration for the stack
// version identified by the create-run URL, without creating a run or mutating
// any state. It feeds the Terraform provider's nuon_stack data source, which
// passes the config to an install-stacks module in place of tfvars. The URL is
// the /v1/stack-runs/{phone_home_id} form the dashboard renders.
func FetchConfig(ctx context.Context, url string) (*Config, error) {
	base, phoneHomeID, err := parseURL(url)
	if err != nil {
		return nil, err
	}
	client := newRunClient(runClientConfig{
		CtlAPIURL:   base,
		PhoneHomeID: phoneHomeID,
	})
	cfg, err := client.fetchConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch stack config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("fetch stack config: ctl-api returned no config block")
	}
	return cfg, nil
}
