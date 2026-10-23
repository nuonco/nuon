package require

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func NotFederatedToken(ctx context.Context, action string) error {
	if !cctx.IsFederatedToken(ctx) {
		return nil
	}

	return stderr.ErrAuthorization{
		Err:         fmt.Errorf("federated tokens cannot %s", action),
		Description: fmt.Sprintf("Tokens issued by an OIDC trust policy cannot %s. Use a user or static API token instead.", action),
	}
}
