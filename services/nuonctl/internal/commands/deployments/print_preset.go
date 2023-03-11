package deployments

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/deployments/presets"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/proto"
)

func (c *commands) PrintPreset(ctx context.Context, componentPreset string) error {
	presetComp, err := presets.New(c.v, componentPreset)
	if err != nil {
		return fmt.Errorf("unable to get preset: %w", err)
	}

	return proto.Print(presetComp)
}
