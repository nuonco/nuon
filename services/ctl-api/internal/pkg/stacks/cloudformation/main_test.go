package cloudformation

import (
	"os"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// Tests serve templates from httptest servers on localhost, which the
// production fetch client refuses by design.
func TestMain(m *testing.M) {
	templateFetchClient = stacks.NewLocalTemplateFetchClient()
	os.Exit(m.Run())
}
