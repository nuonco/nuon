package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/config/sync"
)

type buildResult struct {
	component sync.ComponentState
	buildID   string
	status    string
	success   bool
}

func (s *Service) pollComponentBuilds(ctx context.Context, comps []sync.ComponentState) error {
	if len(comps) == 0 {
		return nil
	}

	cmpByID := make(map[string]sync.ComponentState)
	for _, cmp := range comps {
		cmpByID[cmp.ID] = cmp
	}

	pollTimeout, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	multiSpinner := bubbles.NewMultiSpinnerView()
	for _, cmp := range comps {
		multiSpinner.AddSpinner(cmp.ID, fmt.Sprintf("building component %s %s", cmp.ID, cmp.Name))
	}
	multiSpinner.Start()

	// NOTE: on updates, components are already active and new component_builds records wait to be created.
	time.Sleep(time.Second * 5)

	completedBuilds := make([]buildResult, 0, len(comps))
	var groupError error

	for {
		select {
		case <-pollTimeout.Done():
			err := fmt.Errorf("timeout waiting for components to build")
			ui.PrintError(err)
			for cmpID := range cmpByID {
				cmp := cmpByID[cmpID]
				multiSpinner.CompleteSpinner(cmp.ID, false, "")
				completedBuilds = append(completedBuilds, buildResult{
					component: cmp,
					status:    "timeout",
					success:   false,
				})
			}
			multiSpinner.Stop()
			s.renderSyncResults(ctx, completedBuilds)
			return err
		default:
		}

		completedThisRound := make([]buildResult, 0)

		for cmpID := range cmpByID {
			cmp := cmpByID[cmpID]
			cmpBuild, err := s.api.GetComponentLatestBuild(ctx, cmp.ID)
			if err != nil {
				if nuon.IsServerError(err) {
					multiSpinner.CompleteSpinner(cmp.ID, false, "")
					completedThisRound = append(completedThisRound, buildResult{
						component: cmp,
						status:    "failed",
						success:   false,
					})
					continue
				}
				if nuon.IsNotFound(err) {
					continue
				}
				if cmpBuild == nil {
					continue
				}
			}

			if cmpBuild.Status == componentBuildStatusError {
				multiSpinner.CompleteSpinner(cmp.ID, false, "")
				completedThisRound = append(completedThisRound, buildResult{
					component: cmp,
					buildID:   cmpBuild.ID,
					status:    "failed",
					success:   false,
				})
				groupError = errors.New("at least one build failed")
				continue
			}

			if cmpBuild.Status == componentBuildStatusPolicyFailed {
				multiSpinner.CompleteSpinner(cmp.ID, false, "")
				completedThisRound = append(completedThisRound, buildResult{
					component: cmp,
					buildID:   cmpBuild.ID,
					status:    "failed",
					success:   false,
				})
				groupError = errors.New("at least one build failed due to policy violation")
				continue
			}

			if cmpBuild.Status == componentBuildStatusActive {
				multiSpinner.CompleteSpinner(cmp.ID, true, "")
				completedThisRound = append(completedThisRound, buildResult{
					component: cmp,
					buildID:   cmpBuild.ID,
					status:    "built",
					success:   true,
				})
				continue
			}
		}

		for _, result := range completedThisRound {
			delete(cmpByID, result.component.ID)
			completedBuilds = append(completedBuilds, result)
		}

		if len(cmpByID) == 0 {
			multiSpinner.Stop()
			s.renderSyncResults(ctx, completedBuilds)
			return groupError
		}

		time.Sleep(defaultSyncSleep)
	}
}

func (s *Service) renderSyncResults(ctx context.Context, builds []buildResult) {
	resultsView := bubbles.NewSyncResultsView()

	for _, build := range builds {
		result := bubbles.ComponentResult{
			ID:      build.component.ID,
			Name:    build.component.Name,
			BuildID: build.buildID,
			Status:  build.status,
			Success: build.success,
		}

		if build.buildID != "" {
			reports, err := s.api.GetPolicyReports(ctx, &nuon.PolicyReportsQuery{
				OwnerType: "component_builds",
				OwnerID:   build.buildID,
				Limit:     1,
			})
			if err == nil && len(reports) > 0 {
				report := reports[0]
				result.DenyCount = report.DenyCount
				result.WarnCount = report.WarnCount
				result.PassCount = report.PassCount
				result.PolicyReport = report.ID
			}
		}

		resultsView.AddResult(result)
	}

	ui.PrintRaw(resultsView.Render())
}
