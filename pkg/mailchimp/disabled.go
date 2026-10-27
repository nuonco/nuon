package mailchimp

import "context"

type disabledClient struct{}

var _ Client = (*disabledClient)(nil)

func NewDisabled() Client {
	return &disabledClient{}
}

func (c *disabledClient) Enabled() bool {
	return false
}

func (c *disabledClient) UpsertListMember(_ context.Context, _ string) error {
	return nil
}
