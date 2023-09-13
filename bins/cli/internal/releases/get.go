package releases

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, releaseID string, asJSON bool) {
	view := ui.NewGetView()

	release, err := s.api.GetRelease(ctx, releaseID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON == true {
		j, _ := json.Marshal(release)
		fmt.Println(string(j))
	} else {
		view.Render([][]string{
			[]string{"id", release.ID},
			[]string{"status", release.Status},
			[]string{"created at", release.CreatedAt},
			[]string{"updated at", release.UpdatedAt},
			[]string{"created by", release.CreatedByID},
			[]string{"build id", release.BuildID},
			[]string{"total steps", strconv.Itoa(int(release.TotalReleaseSteps))},
		})
	}
}
