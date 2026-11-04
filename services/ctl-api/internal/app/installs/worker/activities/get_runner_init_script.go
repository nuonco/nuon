package activities

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type GetRunnerInitScriptRequest struct {
	URL string `json:"url" validate:"required"`
}

func FetchRunnerInitScriptRaw(ctx context.Context, url string) ([]byte, error) {
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request for runner init script")
	}

	ctx, cancel := context.WithTimeout(ctx, ghApiTimeout)
	defer cancel()

	r = r.WithContext(ctx)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch runner init script")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unable to fetch runner init script from %s: HTTP %d", url, resp.StatusCode)
	}

	byts, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body of runner init script")
	}

	return byts, nil
}

// @temporal-gen-v2 activity
func (a *Activities) GetRunnerInitScriptRaw(ctx context.Context, req *GetRunnerInitScriptRequest) ([]byte, error) {
	return FetchRunnerInitScriptRaw(ctx, req.URL)
}
