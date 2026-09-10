package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponentBuildAfterQuerySetsIsPreview(t *testing.T) {
	tests := map[string]struct {
		run  *AppBranchRun
		want bool
	}{
		"no branch run": {},
		"standard branch run": {
			run: &AppBranchRun{RunType: AppBranchRunTypeGit},
		},
		"preview record": {
			run:  &AppBranchRun{Preview: &AppBranchRunPreview{}},
			want: true,
		},
		"legacy preview run": {
			run:  &AppBranchRun{RunType: AppBranchRunTypeGitPreview},
			want: true,
		},
		"legacy plan only run": {
			run:  &AppBranchRun{PlanOnly: true},
			want: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			build := ComponentBuild{AppBranchRun: test.run}

			require.NoError(t, build.AfterQuery(nil))
			require.Equal(t, test.want, build.IsPreview)
		})
	}
}
