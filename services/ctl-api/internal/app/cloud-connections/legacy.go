package cloudconnections

import (
	"errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

var ErrLegacyImageAuthDisabled = stderr.NewInvalidRequest(errors.New("legacy image registry authentication is disabled; add the repository in Settings > Image repositories and use a Cloud Connection"))

func legacyImageAuth(enabled bool, disabled error) (*app.CloudConnection, error) {
	if !enabled {
		return nil, disabled
	}
	return nil, nil
}
