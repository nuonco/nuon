package apps

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-go"

	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/config/sync"
)

// BuildOutcome describes the terminal state of one scheduled component build.
type BuildOutcome struct {
	ComponentID   string `json:"component_id"`
	ComponentName string `json:"component_name"`
	Status        string `json:"status"`
}

const (
	buildOutcomeBuilt        = "built"
	buildOutcomeError        = "error"
	buildOutcomePolicyFailed = "policy_failed"
	buildOutcomeTimeout      = "timeout"
	buildOutcomeUnknown      = "unknown"
)

// pollComponentBuilds waits for the scheduled component builds to reach a
// terminal state and returns one outcome per component (in input order). The
// returned error is non-nil when any build did not complete successfully.
func (s *Service) pollComponentBuilds(ctx context.Context, comps []sync.ComponentState) ([]BuildOutcome, error) {
	if len(comps) == 0 {
		return nil, nil
	}

	cmpByID := make(map[string]sync.ComponentState)
	for _, cmp := range comps {
		cmpByID[cmp.ID] = cmp
	}
	statusByID := make(map[string]string)

	pollTimeout, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	multiSpinner := bubbles.NewMultiSpinnerView(s.cfg.Interactive)

	// Add all spinners first
	for _, cmp := range comps {
		multiSpinner.AddSpinner(cmp.ID, fmt.Sprintf("building component %s %s", cmp.ID, cmp.Name))
	}

	// Then start the display
	multiSpinner.Start()

	// NOTE: on updates, components are already active and new component_builds records wait to be created.
	// So we need to wait for the new component_builds to be created before we start to poll.
	time.Sleep(time.Second * 5)

poll:
	for {
		select {
		case <-pollTimeout.Done():
			for cmpID := range cmpByID {
				cmp := cmpByID[cmpID]
				statusByID[cmp.ID] = buildOutcomeTimeout
				multiSpinner.CompleteSpinner(cmp.ID, false, fmt.Sprintf("timeout waiting for component %s %s to build", cmp.ID, cmp.Name))
			}
			break poll
		default:
		}

		completedComponents := make([]sync.ComponentState, 0)

		for cmpID := range cmpByID {
			cmp := cmpByID[cmpID]
			cmpBuild, err := s.api.GetComponentLatestBuild(ctx, cmp.ID)
			if err != nil {
				if nuon.IsServerError(err) {
					statusByID[cmp.ID] = buildOutcomeUnknown
					multiSpinner.CompleteSpinner(cmp.ID, false, fmt.Sprintf("unable to check build status for component %s %s", cmp.ID, cmp.Name))
					completedComponents = append(completedComponents, cmp)
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
				statusByID[cmp.ID] = buildOutcomeError
				multiSpinner.CompleteSpinner(cmp.ID, false, fmt.Sprintf("error building component %s %s", cmp.ID, cmp.Name))
				completedComponents = append(completedComponents, cmp)
				continue
			}
			if cmpBuild.Status == componentBuildStatusPolicyFailed {
				statusByID[cmp.ID] = buildOutcomePolicyFailed
				multiSpinner.CompleteSpinner(cmp.ID, false, fmt.Sprintf("policy violation for component %s %s", cmp.ID, cmp.Name))
				completedComponents = append(completedComponents, cmp)
				continue
			}

			if cmpBuild.Status == componentBuildStatusActive {
				statusByID[cmp.ID] = buildOutcomeBuilt
				multiSpinner.CompleteSpinner(cmp.ID, true, fmt.Sprintf("finished building component %s %s", cmp.ID, cmp.Name))
				completedComponents = append(completedComponents, cmp)
				continue
			}
		}

		// Remove completed components from tracking
		for _, cmp := range completedComponents {
			delete(cmpByID, cmp.ID)
		}

		if len(cmpByID) == 0 {
			break poll
		}

		time.Sleep(defaultSyncSleep)
	}

	multiSpinner.Stop()

	outcomes := make([]BuildOutcome, 0, len(comps))
	for _, cmp := range comps {
		outcomes = append(outcomes, BuildOutcome{
			ComponentID:   cmp.ID,
			ComponentName: cmp.Name,
			Status:        statusByID[cmp.ID],
		})
	}

	return outcomes, buildOutcomesErr(outcomes)
}

// buildOutcomesErr summarizes non-successful outcomes into a single error, or
// nil when every build completed.
func buildOutcomesErr(outcomes []BuildOutcome) error {
	failures := make([]string, 0)
	for _, o := range outcomes {
		if o.Status == buildOutcomeBuilt {
			continue
		}
		failures = append(failures, fmt.Sprintf("%s (%s)", o.ComponentName, o.Status))
	}
	if len(failures) == 0 {
		return nil
	}

	return fmt.Errorf("%d of %d component build(s) did not complete: %s",
		len(failures), len(outcomes), strings.Join(failures, ", "))
}
