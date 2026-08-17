package workflows

import (
	"reflect"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// Activities are registered per worker, so an activity scheduled from more than
// one namespace has to stay in the shared set every worker registers. When it
// does not, the caller fails with ActivityNotRegisteredError — which callers
// here handle by falling back, so the breakage is silent.
func TestSharedActivitySet(t *testing.T) {
	// GetSandboxBuildOCIRegistry: scheduled from the apps namespace (building the
	// sandbox artifact) and the installs namespace (resolving it when planning a
	// sandbox run). Missing, the installs worker silently falls back to git.
	//
	// GetGARAccessToken: scheduled from the components namespace (pulling a
	// source image) and the installs namespace (authenticating the sandbox
	// artifact pull for a runner outside GCP).
	methods := []string{
		"GetSandboxBuildOCIRegistry",
		"GetGARAccessToken",
	}

	shared := &Activities{Activities: &activities.Activities{}}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			for _, acts := range shared.AllActivities() {
				if acts == nil {
					continue
				}
				v := reflect.ValueOf(acts)
				if v.Kind() == reflect.Ptr && v.IsNil() {
					continue
				}
				if v.MethodByName(method).IsValid() {
					return
				}
			}

			t.Fatalf("%s is not reachable through Activities.AllActivities(); "+
				"workers that do not register the owning namespace cannot run it", method)
		})
	}
}
