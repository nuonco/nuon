package activities

import (
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type GetPhoneHomeScriptRequest struct{}

// @temporal-gen activity
func (a *Activities) GetPhoneHomeScriptRaw(ctx context.Context, req *GetPhoneHomeScriptRequest) ([]byte, error) {

	// Grab the latest version of the phone-home script
	resp, err := http.Get("https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/phonehome.py")
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch phone-home script")
	}
	byts, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body of phone-home script")
	}

	return byts, nil
}
