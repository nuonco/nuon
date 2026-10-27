package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const phoneHomeTimeout = 10 * time.Second

// sandboxPhoneHomeTokenTTL is long enough for the activity + a few retries, short
// enough that a leaked sandbox token is useless outside the provision window.
const sandboxPhoneHomeTokenTTL = 15 * time.Minute

type FireSandboxPhoneHomeRequest struct {
	InstallID   string         `validate:"required"`
	PhoneHomeID string         `validate:"required"`
	Data        map[string]any `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) FireSandboxPhoneHome(ctx context.Context, req *FireSandboxPhoneHomeRequest) error {
	var install app.Install
	if res := a.db.WithContext(ctx).
		Where(app.Install{ID: req.InstallID}).
		First(&install); res.Error != nil {
		return errors.Wrap(res.Error, "unable to load install")
	}

	var stack app.InstallStack
	if res := a.db.WithContext(ctx).
		Where(app.InstallStack{InstallID: install.ID}).
		First(&stack); res.Error != nil {
		return errors.Wrap(res.Error, "unable to load install stack")
	}

	token, err := a.acctClient.IssueToken(ctx, account.ServiceAccountEmail(stack.ID), sandboxPhoneHomeTokenTTL, account.IssueTokenRequest{
		OrgID: install.OrgID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to mint stack service account token")
	}

	base := a.cfg.StackAPIURL
	if base == "" {
		base = fmt.Sprintf("http://127.0.0.1:%s", a.cfg.StackHTTPPort)
	}
	url := fmt.Sprintf("%s/v1/stacks/%s/phone-home",
		strings.TrimSuffix(base, "/"), install.ID)

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
	r.Header.Set("Authorization", "Bearer "+token.Token)
	r.Header.Set("X-Nuon-Org-ID", install.OrgID)
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
