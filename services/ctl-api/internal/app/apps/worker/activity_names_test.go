package worker

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	appconfigsync "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/appconfigsync"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	syncappconfiginstalls "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/syncappconfiginstalls"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
)

// Temporal registers activities by method name, so a name used by two of the
// activity sets on this worker panics the whole worker at startup — every signal
// on every apps queue then stops processing.
func TestActivityNamesAreUniqueAcrossWorker(t *testing.T) {
	sets := []any{
		(*activities.Activities)(nil),
		(*branchactivities.Activities)(nil),
		(*syncappconfiginstalls.Activities)(nil),
		(*appconfigsync.Activities)(nil),
	}

	owner := map[string]string{}
	for _, set := range sets {
		typ := reflect.TypeOf(set)
		for i := range typ.NumMethod() {
			name := typ.Method(i).Name
			if prev, dup := owner[name]; dup {
				assert.Failf(t, "duplicate activity name",
					"%q is registered by both %s and %s", name, prev, typ.String())
				continue
			}
			owner[name] = typ.String()
		}
	}
}
