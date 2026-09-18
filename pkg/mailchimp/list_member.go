package mailchimp

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultAPITimeout time.Duration = time.Second * 10

type upsertListMemberRequest struct {
	EmailAddress string `json:"email_address"`
	StatusIfNew  string `json:"status_if_new"`
}

// UpsertListMember adds an email to the configured audience, or leaves an existing member's
// status untouched: status_if_new only applies on insert, so it never resubscribes someone
// who unsubscribed in Mailchimp.
func (c *client) UpsertListMember(ctx context.Context, email string) error {
	byts, err := json.Marshal(upsertListMemberRequest{
		EmailAddress: email,
		StatusIfNew:  "subscribed",
	})
	if err != nil {
		return fmt.Errorf("unable to create request json: %w", err)
	}

	url := fmt.Sprintf("%s/lists/%s/members/%s", c.baseURL, c.AudienceID, subscriberHash(email))
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(byts))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	req.SetBasicAuth("nuon", c.APIKey)
	req.Header.Add("Content-Type", "application/json")

	timeoutCtx, cancelFn := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancelFn()
	req = req.WithContext(timeoutCtx)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("response error: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("unexpected response status %d: %s", res.StatusCode, string(body))
	}

	return nil
}

// subscriberHash is Mailchimp's member identifier: the md5 of the lowercased email.
func subscriberHash(email string) string {
	sum := md5.Sum([]byte(strings.ToLower(email)))
	return hex.EncodeToString(sum[:])
}
