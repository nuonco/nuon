package controlplanejob

import (
	"context"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func (a *Activities) injectInstallCredentials(_ context.Context, _ string, _ *plantypes.CompositePlan) error {
	return nil
}
