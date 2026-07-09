package workflow

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestEmptyStepsViewReportsStatus(t *testing.T) {
	testCases := []struct {
		name     string
		status   models.AppStatus
		contains string
	}{
		{"in progress", models.AppStatusInDashProgress, "no steps yet"},
		{"success", models.AppStatusSuccess, "finished with no steps"},
		{"error", models.AppStatusError, "finished with no steps"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := model{
				stepDetail: viewport.New(viewport.WithWidth(100), viewport.WithHeight(20)),
				workflow: &models.AppWorkflow{
					Status: &models.AppCompositeStatus{Status: tc.status},
				},
			}

			out := m.emptyStepsView()
			if strings.Contains(out, "Loading") || strings.Contains(out, "loading") {
				t.Fatalf("empty-steps view must not show a loading state, got:\n%s", out)
			}
			if !strings.Contains(out, tc.contains) {
				t.Fatalf("expected %q in empty-steps view, got:\n%s", tc.contains, out)
			}
		})
	}
}

// TestPopulateStepDetailViewNoLongerSpinsWhenFetched guards the reported bug:
// once a workflow has been fetched, a zero-step workflow must render a terminal
// message rather than "Loading ..." forever.
func TestPopulateStepDetailViewNoLongerSpinsWhenFetched(t *testing.T) {
	m := &model{
		stepDetail: viewport.New(viewport.WithWidth(100), viewport.WithHeight(20)),
		workflow: &models.AppWorkflow{
			Status: &models.AppCompositeStatus{Status: models.AppStatusInDashProgress},
		},
	}

	m.populateStepDetailView(false)
	if strings.Contains(m.stepDetail.View(), "Loading ...") {
		t.Fatal("fetched zero-step workflow must not render the loading placeholder")
	}

	// no workflow yet: loading is still the correct state
	loadingModel := &model{
		stepDetail: viewport.New(viewport.WithWidth(100), viewport.WithHeight(20)),
	}
	loadingModel.populateStepDetailView(false)
	if !strings.Contains(loadingModel.stepDetail.View(), "Loading") {
		t.Fatal("pre-fetch state should still show loading")
	}
}
