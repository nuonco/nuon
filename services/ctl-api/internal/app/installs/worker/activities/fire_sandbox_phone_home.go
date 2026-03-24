package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const phoneHomeTimeout = 10 * time.Second

type FireSandboxPhoneHomeRequest struct {
	InstallID   string         `validate:"required"`
	PhoneHomeID string         `validate:"required"`
	Data        map[string]any `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) FireSandboxPhoneHome(ctx context.Context, req *FireSandboxPhoneHomeRequest) error {
	url := fmt.Sprintf("%s/v1/installs/%s/phone-home/%s", a.cfg.PublicAPIURL, req.InstallID, req.PhoneHomeID)

	data := req.Data
	data["request_type"] = "Create"

	body, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "unable to marshal phone home request")
	}

	ctx, cancel := context.WithTimeout(ctx, phoneHomeTimeout)
	defer cancel()

	r, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "unable to create phone home request")
	}
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(ctx)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return errors.Wrap(err, "unable to fire phone home request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("phone home returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
