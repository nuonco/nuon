package cloudformation

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultAWSPhoneHomeScript = "https://raw.githubusercontent.com/nuonco/runner/refs/tags/aws-v0.1.4/scripts/aws/phonehome.py"

const phoneHomeScriptFetchTimeout = 5 * time.Second

func FetchPhoneHomeScript(ctx context.Context, appURL, environmentURL string) ([]byte, error) {
	url := DefaultAWSPhoneHomeScript
	if appURL != "" {
		url = appURL
	} else if environmentURL != "" {
		url = environmentURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for phone-home script: %w", err)
	}

	client := &http.Client{Timeout: phoneHomeScriptFetchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch phone-home script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unable to fetch phone-home script from %s: HTTP %d", url, resp.StatusCode)
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read phone-home script: %w", err)
	}
	return contents, nil
}
