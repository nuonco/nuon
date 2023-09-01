package version

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Version(ctx context.Context) error {
	ui.Line(ctx, "%s", "development")
	return nil
}
