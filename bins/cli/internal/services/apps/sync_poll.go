package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-go"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
	"github.com/powertoolsdev/mono/pkg/config/sync"
)

func (s *Service) pollComponentBuilds(ctx context.Context, comps []sync.ComponentState) error {
	// Early return if no components to build
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
	
	// Add all spinners first
	for _, cmp := range comps {
		multiSpinner.AddSpinner(cmp.ID, fmt.Sprintf("building component %s %s", cmp.ID, cmp.Name))
	}
	
	// Then start the display
	multiSpinner.Start()

	// NOTE: on updates, components are already active and new component_builds records wait to be created.
	// So we need to wait for the new component_builds to be created before we start to poll.
	time.Sleep(time.Second * 5)

	for {
		select {
		case <-pollTimeout.Done():
			err := fmt.Errorf("timeout waiting for components to build")
			ui.PrintError(err)
			for cmpID := range cmpByID {
				cmp := cmpByID[cmpID]
				multiSpinner.CompleteSpinner(cmp.ID, false, fmt.Sprintf("timeout waiting for component %s %s to build", cmp.ID, cmp.Name))
			}
			multiSpinner.Stop()
			return err
		default:
		}

		var groupError error
		completedComponents := make([]string, 0)

		for cmpID := range cmpByID {
			cmp := cmpByID[cmpID]
			cmpBuild, err := s.api.GetComponentLatestBuild(ctx, cmp.ID)
			if err != nil {
				if nuon.IsServerError(err) {
					multiSpinner.CompleteSpinner(cmp.ID, false, fmt.Sprintf("error building component %s %s", cmp.ID, cmp.Name))
					completedComponents = append(completedComponents, cmpID)
					continue
				}
				// in case we didn't wait long enough for an initial build record, ignore and loop again
				if nuon.IsNotFound(err) {
					continue
				}
				// TODO: avoid panic if we error on network issues. We should introduce a retryer at the sdk level.
				// for now, this loop is inherently retrying.
				if cmpBuild == nil {
					continue
				}
			}
			if cmpBuild.Status == componentBuildStatusError {
				multiSpinner.CompleteSpinner(cmp.ID, false, fmt.Sprintf("error building component %s %s", cmp.ID, cmp.Name))
				completedComponents = append(completedComponents, cmpID)
				groupError = errors.New("at least one build failed")
				continue
			}

			if cmpBuild.Status == componentBuildStatusActive {
				multiSpinner.CompleteSpinner(cmp.ID, true, fmt.Sprintf("finished building component %s %s", cmp.ID, cmp.Name))
				completedComponents = append(completedComponents, cmpID)
				continue
			}
		}

		// Remove completed components from tracking
		for _, cmpID := range completedComponents {
			delete(cmpByID, cmpID)
		}

		if len(cmpByID) == 0 {
			multiSpinner.Stop()
			return groupError
		}

		time.Sleep(defaultSyncSleep)
	}
}
