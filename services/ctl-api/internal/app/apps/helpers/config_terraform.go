package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) GenerateTerraformJSONFromConfig(ctx context.Context, inputCfg []byte, configFmt app.AppConfigFmt) ([]byte, error) {
	tfJSON, err := parse.ToTerraformJSON(parse.ParseConfig{
		V:           h.v,
		BackendType: config.BackendTypeS3,
		Template:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get terraform json: %w", err)
	}

	return tfJSON, nil
}
