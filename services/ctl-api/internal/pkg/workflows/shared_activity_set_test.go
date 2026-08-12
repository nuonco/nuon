package workflows

import (
	"reflect"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// GetSandboxBuildOCIRegistry is scheduled from both the apps namespace (building
// the sandbox artifact) and the installs namespace (resolving it when planning a
// sandbox run). Activities are registered per worker, so it has to stay in the
// shared set every worker registers — otherwise the installs worker fails the
// call with ActivityNotRegisteredError and silently falls back to a git source.
func TestGetSandboxBuildOCIRegistryIsInSharedActivitySet(t *testing.T) {
	shared := &Activities{Activities: &activities.Activities{}}

	const method = "GetSandboxBuildOCIRegistry"
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

	t.Fatalf("%s is not reachable through Activities.AllActivities(); workers that "+
		"do not register the apps branch activities cannot run it", method)
}
