package activities

import (
	"context"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
)

type GetGARAccessTokenRequest struct {
	ServiceAccountEmail      string
	WorkloadIdentityProvider string
}

type GARAccessToken struct {
	Username string
	Password string
}

// GetGARAccessToken lives in the shared activity set because both the components
// namespace (pulling a source image for a build) and the installs namespace
// (resolving the sandbox artifact) schedule it. Activities are registered per
// worker, so a components-only registration is invisible to installs.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetGARAccessToken(ctx context.Context, req *GetGARAccessTokenRequest) (*GARAccessToken, error) {
	token, err := a.registryAuth.GAR(ctx, temporalzap.GetActivityLogger(ctx), req.ServiceAccountEmail, req.WorkloadIdentityProvider)
	if err != nil {
		return nil, err
	}

	return &GARAccessToken{Username: token.Username, Password: token.Password}, nil
}
