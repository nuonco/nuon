package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, compID string) {
	basicText := ui.NewBasicText()

	builds, err := s.api.GetComponentBuilds(ctx, compID)
	if err != nil {
		basicText.PrintOnError(err)
		return
	}

	if len(builds) == 0 {
		basicText.Println("No builds found")
	} else {
		for _, build := range builds {
			basicText.Printfln("%s - %s", build.ID, build.Status)
		}
	}

}
