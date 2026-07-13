package stack

import (
	"context"
	"fmt"
)

// FetchConfig reads the rendered install-stack config for the stack version
// identified by the create-run URL.
func FetchConfig(ctx context.Context, url string) (*Config, error) {
	base, phoneHomeID, err := parseURL(url)
	if err != nil {
		return nil, err
	}
	client := newRunClient(runClientConfig{
		RunnerAPIURL: base,
		PhoneHomeID:  phoneHomeID,
	})
	cfg, err := client.fetchConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch stack config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("fetch stack config: runner api returned no config block")
	}
	return cfg, nil
}
