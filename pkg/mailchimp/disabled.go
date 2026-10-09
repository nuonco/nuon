package mailchimp

import "context"

type disabledClient struct{}

var _ Client = (*disabledClient)(nil)

// NewDisabled returns a no-op client for deployments without Mailchimp configuration.
func NewDisabled() Client {
	return &disabledClient{}
}

func (c *disabledClient) UpsertListMember(_ context.Context, _ string) error {
	return nil
}
