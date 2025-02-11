package activities

import (
	"context"

	"github.com/pkg/errors"
)

type DeleteRequest struct {
	OrgID string `validate:"required"`
}

// @temporal-gen activity
// @by-id OrgID
func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
	if err := a.helpers.HardDelete(ctx, req.OrgID); err != nil {
		return errors.Wrap(err, "unable to delete org")
	}

	return nil
}
