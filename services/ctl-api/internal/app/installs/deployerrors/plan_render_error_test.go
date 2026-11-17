package deployerrors

import (
	"testing"
)

func TestDeployPlanRenderError_HowToFix(t *testing.T) {
	cases := []struct {
		name            string
		appSyncDisabled bool
		want            string
	}{
		{"app sync enabled", false, "Check the referenced value exists in your app config — a missing sandbox output, action workflow output, or input is the usual cause. Fix the reference, re-run `nuon apps sync` or run your app branch with the fix, then retry the deploy."},
		{"app sync disabled", true, "Check the referenced value exists in your app config — a missing sandbox output, action workflow output, or input is the usual cause. Fix the reference, run your app branch with the fix, then retry the deploy."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &DeployPlanRenderError{ComponentName: "api", AppSyncDisabled: tc.appSyncDisabled}
			sections := e.Sections()
			last := sections[len(sections)-1]
			if last.Heading != "How to fix" || last.Body != tc.want {
				t.Errorf("last section = %q: %q, want How to fix: %q", last.Heading, last.Body, tc.want)
			}
		})
	}
}
