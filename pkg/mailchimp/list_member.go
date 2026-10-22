package mailchimp

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	mailchimp "github.com/mailchimp/mailchimp-marketing-go-sdk"
)

const defaultAPITimeout time.Duration = time.Second * 10

// UpsertListMember adds an email to the configured audience, or leaves an existing member's
// status untouched: status_if_new only applies on insert, so it never resubscribes someone
// who unsubscribed in Mailchimp.
func (c *client) UpsertListMember(ctx context.Context, email string) error {
	timeoutCtx, cancelFn := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancelFn()

	_, err := c.sdk.Lists.UpsertMember(timeoutCtx, &mailchimp.UpsertMemberListsRequest{
		ListID:         c.AudienceID,
		SubscriberHash: subscriberHash(email),
		EmailAddress:   email,
		StatusIfNew:    mailchimp.UpsertMemberListsRequestStatusIfNewSubscribed.Ptr(),
	})
	if err != nil {
		return fmt.Errorf("unable to upsert list member: %w", err)
	}

	return nil
}

// subscriberHash is Mailchimp's member identifier: the md5 of the lowercased email.
func subscriberHash(email string) string {
	sum := md5.Sum([]byte(strings.ToLower(email)))
	return hex.EncodeToString(sum[:])
}
