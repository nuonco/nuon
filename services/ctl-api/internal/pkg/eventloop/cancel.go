package eventloop

import (
	"context"

	"github.com/pkg/errors"
)

func (a *evClient) Cancel(ctx context.Context, namespace, id string) error {
	err := a.client.CancelWorkflowInNamespace(ctx, namespace, id, "")
	if err != nil {
		return errors.Wrap(err, "unable to cancel workflow")
	}

	return nil
}
