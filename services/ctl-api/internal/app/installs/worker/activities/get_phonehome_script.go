package activities

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	ghApiTimeout = 5 * time.Second
)

type GetPhoneHomeScriptRequest struct{}

// @temporal-gen activity
func (a *Activities) GetPhoneHomeScriptRaw(ctx context.Context, req *GetPhoneHomeScriptRequest) ([]byte, error) {

	r, err := http.NewRequest(http.MethodGet, "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/phonehome.py", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request for phone-home script")
	}

	ctx, cancel := context.WithTimeout(ctx, ghApiTimeout)
	defer cancel()

	r = r.WithContext(ctx)
	client := http.DefaultClient

	// Grab the latest version of the phone-home script
	resp, err := client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch phone-home script")
	}
	byts, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body of phone-home script")
	}

	return byts, nil
}
