package dev

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-go"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config/sync"
)

func (s *Service) pollComponentBuilds(ctx context.Context, comps []sync.ComponentState) error {
	cmpByID := make(map[string]sync.ComponentState)
	for _, cmp := range comps {
		cmpByID[cmp.ID] = cmp
	}

	pollTimeout, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	multi := pterm.DefaultMultiPrinter

	spinnersByComponentID := make(map[string]*pterm.SpinnerPrinter)
	for _, cmp := range comps {
		spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start(fmt.Sprintf("building component %s %s", cmp.ID, cmp.Name))
		spinnersByComponentID[cmp.ID] = spinner
	}

	multi.Start()

	// NOTE: on updates, components are already active and new component_builds records wait to be created.
	// So we need to wait for the new component_builds to be created before we start to poll.
	time.Sleep(time.Second * 5)

	for {
		select {
		case <-pollTimeout.Done():
			err := fmt.Errorf("timeout waiting for components to build")
			ui.PrintError(err)
			for cmpID, spinner := range spinnersByComponentID {
				cmp, _ := cmpByID[cmpID]
				spinner.Fail(fmt.Sprintf("timeout waiting for component %s %s to build", cmp.ID, cmp.Name))
			}
			multi.Stop()
			return err
		default:
		}

		var err error = nil
		for cmpID := range spinnersByComponentID {
			cmp, _ := cmpByID[cmpID]
			cmpBuild, err := s.api.GetComponentLatestBuild(ctx, cmp.ID)
			if err != nil {
				if nuon.IsServerError(err) {
					spinnersByComponentID[cmpID].Fail(fmt.Sprintf("error building component %s %s", cmp.ID, cmp.Name))
					delete(spinnersByComponentID, cmpID)
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
				spinnersByComponentID[cmpID].Fail(fmt.Sprintf("error building component %s %s", cmp.ID, cmp.Name))
				delete(spinnersByComponentID, cmpID)
				err = errors.New("at least one build failed")
				continue
			}

			if cmpBuild.Status == componentBuildStatusActive {
				spinnersByComponentID[cmpID].Success(fmt.Sprintf("finished building component %s %s", cmp.ID, cmp.Name))
				delete(spinnersByComponentID, cmpID)
				continue
			}
		}

		if len(spinnersByComponentID) == 0 {
			multi.Stop()
			return err
		}

		time.Sleep(defaultSyncSleep)
	}
}
