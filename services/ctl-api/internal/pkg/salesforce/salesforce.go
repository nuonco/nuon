package salesforce

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

const (
	TrialSignupSubject = "trial-signup"

	defaultTimeout = time.Second * 5
)

type TrialSignup struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	CompanyName string `json:"companyName"`
	JobTitle    string `json:"jobTitle"`
	Notes       string `json:"notes"`
	Subject     string `json:"subject"`
}

type Client interface {
	Enabled() bool
	SendTrialSignup(ctx context.Context, signup TrialSignup) error
}

type client struct {
	endpoint string
}

var _ Client = (*client)(nil)

func New(cfg *internal.Config) Client {
	return &client{
		endpoint: cfg.SFTrialEndpoint,
	}
}

func (c *client) Enabled() bool {
	return c.endpoint != ""
}

func (c *client) SendTrialSignup(ctx context.Context, signup TrialSignup) error {
	if !c.Enabled() {
		return nil
	}

	byts, err := json.Marshal(signup)
	if err != nil {
		return fmt.Errorf("unable to marshal trial signup: %w", err)
	}

	timeoutCtx, cancelFn := context.WithTimeout(ctx, defaultTimeout)
	defer cancelFn()

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodPost, c.endpoint, bytes.NewReader(byts))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send trial signup: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("trial signup request failed with status %d: %s", res.StatusCode, string(body))
	}

	return nil
}
