package sender

import "context"

type noopNotifier struct{}

// NewNoopSender instantiates a sender that does nothing
func NewNoopSender() *noopNotifier {
	return &noopNotifier{}
}

// Send does nothing
// Good for testing or running locally
func (n *noopNotifier) Send(ctx context.Context, msg string) error {
	return nil
}
