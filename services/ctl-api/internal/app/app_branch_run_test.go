package app

import "testing"

// Only the VCS push path writes run_type into workflow metadata, so a plan-only
// run triggered through the API used to fall through to the deploy path and
// start real install workflows.
func TestAppBranchRunIsPreview(t *testing.T) {
	cases := []struct {
		name string
		run  AppBranchRun
		want bool
	}{
		{"pr preview", AppBranchRun{RunType: AppBranchRunTypeGitPreview}, true},
		{"plan-only manual", AppBranchRun{RunType: AppBranchRunTypeManual, PlanOnly: true}, true},
		{"plan-only git push", AppBranchRun{RunType: AppBranchRunTypeGit, PlanOnly: true}, true},
		{"manual deploy", AppBranchRun{RunType: AppBranchRunTypeManual}, false},
		{"git push deploy", AppBranchRun{RunType: AppBranchRunTypeGit}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.run.IsPreview(); got != tc.want {
				t.Fatalf("IsPreview() = %v, want %v", got, tc.want)
			}
		})
	}
}
