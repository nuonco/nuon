package releases

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Steps(ctx context.Context, releaseID string) error {
	steps, err := s.api.GetReleaseSteps(ctx, releaseID)
	if err != nil {
		return err
	}

	if len(steps) == 0 {
		ui.Line(ctx, "No components found")
	} else {
		for _, step := range steps {
			ui.Line(ctx, "%s - %s", step.ID, step.Status)
		}
	}

	return nil
}
