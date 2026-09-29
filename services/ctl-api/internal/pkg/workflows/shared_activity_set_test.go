package workflows

import (
	"reflect"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func TestSharedActivitySet(t *testing.T) {
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
