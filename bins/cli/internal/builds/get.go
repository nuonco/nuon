package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, compID, buildID string) {
	basicText := ui.NewBasicText()

	build, err := s.api.GetComponentBuild(ctx, compID, buildID)
	if err != nil {
		basicText.PrintOnError(err)
		return
	}

	basicText.Printfln("%s - %s", build.ID, build.Status)
}
